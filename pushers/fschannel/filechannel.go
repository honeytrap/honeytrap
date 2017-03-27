package fschannel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var crtlline = []byte("\r\n")
var defaultMaxSize = 1024 * 1024 * 1024
var defaultWaitTime = 5 * time.Second
var log = logging.MustGetLogger("honeytrap:channels:elasticsearch")

// FileChannel defines a struct which implements the pushers.Pusher interface
// and allows us to write PushMessage updates into a giving file path. Mainly for
// the need to sync PushMessage to local files for persistence.
// File paths provided are either created with a append mode if they already
// exists else will be created. FileChannel will also restrict filesize to a max of 1gb by default else if
// there exists a max size set in configuration, then that will be used instead,
// also the old file will be renamed with the current timestamp and a new file created.
type FileChannel struct {
	maxSize  int
	destFile string
	dest     *os.File
	ms       time.Duration
	filters  map[string]*regexp.Regexp
	request  chan *message.PushMessage
	wg       sync.WaitGroup
}

// New returns a new instance of a FileChannel.
func New() *FileChannel {
	var fc FileChannel
	fc.ms = defaultWaitTime
	fc.maxSize = defaultMaxSize
	fc.request = make(chan *message.PushMessage)
	fc.filters = make(map[string]*regexp.Regexp, 0)

	return &fc
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

	for _, message := range messages {
		if matcher, ok := f.filters["sensor"]; ok && !matcher.MatchString(message.Sensor) {
			continue
		}

		if matcher, ok := f.filters["category"]; ok && !matcher.MatchString(message.Category) {
			continue
		}

		if matcher, ok := f.filters["session_id"]; ok && !matcher.MatchString(message.SessionID) {
			continue
		}

		if matcher, ok := f.filters["container_id"]; ok && !matcher.MatchString(message.ContainerID) {
			continue
		}

		f.request <- message
	}

	// Close channel
	close(f.request)
}

// UnmarshalConfig takes a provide configuration map type and sets the
// underline configuration for the giving file channel.
func (f *FileChannel) UnmarshalConfig(c interface{}) error {
	conf, ok := c.(map[string]interface{})
	if !ok {
		return errors.New("Invalid configuration type, expected a map")
	}

	targetFile, ok := conf["file"].(string)
	if !ok {
		return errors.New("Expected 'file' key for target file path")
	}

	if strings.TrimSpace(targetFile) == "" {
		return errors.New("Expected 'file' value not to be empty")
	}

	if filters, ok := conf["filters"].(map[string]interface{}); ok {
		for name, val := range filters {
			name = strings.ToLower(name)
			switch rVal := val.(type) {
			case *regexp.Regexp:
				f.filters[name] = rVal
			case string:
				f.filters[name] = regexp.MustCompile(rVal)
			}
		}
	}

	if waitMS, ok := conf["ms"].(string); ok {
		f.ms = config.MakeDuration(waitMS, int(defaultWaitTime))
	}

	if mxSize, ok := conf["max_size"].(string); ok {
		f.maxSize = config.ConvertToInt(mxSize, defaultMaxSize)
	}

	f.destFile = targetFile

	return nil
}

// syncWrites startups the channel procedure to listen for new writes to giving file.
func (f *FileChannel) syncWrites() error {
	log.Info("FileChannel.syncWrites : Started")

	if f.dest != nil {
		log.Info("FileChannel.syncWrites : Completed : Already Running")
		return nil
	}

	if f.request == nil {
		f.request = make(chan *message.PushMessage)
	}

	var err error

	f.dest, err = newFile(f.destFile, f.maxSize)
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

	ticker := time.NewTimer(f.ms)
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
