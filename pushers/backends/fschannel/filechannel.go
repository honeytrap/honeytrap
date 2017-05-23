package fschannel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var (
	defaultMaxSize  = 1024 * 1024 * 1024
	defaultWaitTime = 5 * time.Second
	crtlline        = []byte("\r\n")
	log             = logging.MustGetLogger("honeytrap:channels:filechannel")
)

// FileConfig defines the config used to setup the FileChannel.
type FileConfig struct {
	MaxSize int    `toml:"maxsize"`
	File    string `toml:"file"`
	Timeout string `toml:"timeout"`
}

// FileChannel defines a struct which implements the pushers.Pusher interface
// and allows us to write PushMessage updates into a giving file path. Mainly for
// the need to sync PushMessage to local files for persistence.
// File paths provided are either created with a append mode if they already
// exists else will be created. FileChannel will also restrict filesize to a max of 1gb by default else if
// there exists a max size set in configuration, then that will be used instead,
// also the old file will be renamed with the current timestamp and a new file created.
type FileChannel struct {
	config  FileConfig
	timeout time.Duration
	dest    *os.File
	request chan message.PushMessage
	wg      sync.WaitGroup
}

// New returns a new instance of a FileChannel.
func New(c FileConfig) *FileChannel {
	var fc FileChannel
	fc.config = c
	fc.request = make(chan message.PushMessage)
	fc.timeout = config.MakeDuration(c.Timeout, int(defaultWaitTime))

	return &fc
}

// NewWith defines a function to return a pushers.Channel which delivers
// new messages to a giving underline system file, defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig FileConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.File == "" {
		return nil, errors.New("fschannel.FileConfig Invalid: File can not be empty")
	}

	return New(apiconfig), nil
}

func init() {
	pushers.RegisterBackend("file", NewWith)
}

// Wait calls the internal waiter.
func (f *FileChannel) Wait() {
	f.wg.Wait()
}

// Send delivers the giving if it passes all filtering criteria into the
// FileChannel write queue.
func (f *FileChannel) Send(messages []message.PushMessage) {
	log.Info("FileChannel.Send : Started")

	if err := f.syncWrites(); err != nil {
		log.Errorf("FileChannel.Send : Completed : %+q", err)
		return
	}

	// Close channel.
	close(f.request)
}

// syncWrites startups the channel procedure to listen for new writes to giving file.
func (f *FileChannel) syncWrites() error {
	log.Info("FileChannel.syncWrites : Started")

	if f.dest != nil {
		log.Info("FileChannel.syncWrites : Completed : Already Running")
		return nil
	}

	if f.request == nil {
		f.request = make(chan message.PushMessage)
	}

	var err error

	f.dest, err = newFile(f.config.File, f.config.MaxSize)
	if err != nil {
		log.Info("FileChannel.syncWrites : Completed : Failed create destination file")
		return err
	}

	f.wg.Add(1)
	go f.syncLoop()

	log.Info("FileChannel.syncWrites : Completed")
	return nil
}

// syncLoop handles configuration of the giving loop for writing to file.
func (f *FileChannel) syncLoop() {
	defer f.wg.Done()

	ticker := time.NewTimer(f.timeout)
	var buf bytes.Buffer

	{
	writeSync:
		for {
			select {
			case <-ticker.C:
				f.dest.Close()
				f.dest = nil

				// Close request channel and nil it.
				close(f.request)
				f.request = nil

				break writeSync

			case req, ok := <-f.request:
				if !ok {
					f.dest.Close()
					f.request = nil
					f.dest = nil
					break writeSync
				}

				if err := json.NewEncoder(&buf).Encode(req); err != nil {
					log.Errorf("FileChannel.syncWrites : Failed to marshal PushMessage to JSON : %+q", err)
					continue writeSync
				}

				// Add the line control to each jsonified message.
				buf.Write(crtlline)

				if _, err := io.Copy(f.dest, &buf); err != nil && err != io.EOF {
					log.Errorf("FileChannel.syncWrites : Failed to copy data to File : %+q", err)
				}

				if err := f.dest.Sync(); err != nil {
					log.Errorf("FileChannel.syncWrites : Failed to sync Write to File : %+q", err)
				}

				// Reset the buffer for reuse.
				buf.Reset()
			}
		}
	}
}

// newFile returns a new file with the giving target path and returns the
// new file object.
func newFile(targetPath string, maxSize int) (*os.File, error) {

	// Attempt to stat file, if it does not exists then create a new one.
	stat, err := os.Stat(targetPath)
	if err != nil {
		dest, err := os.Create(targetPath)
		if err != nil {
			return nil, err
		}

		return dest, nil
	}

	if stat.IsDir() {
		return nil, errors.New("Only direct file paths allowed")
	}

	// if we dealing with a file still  below our max size, then
	// open file if already exists else
	if int(stat.Size()) <= maxSize {
		dest, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}

		return dest, nil
	}

	if err := os.Rename(targetPath, fmt.Sprintf("%s-%s", targetPath, stat.ModTime().UTC())); err != nil {
		return nil, err
	}

	return os.Create(targetPath)
}
