package fschannel_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/honeytrap/honeytrap/pushers/backends/fschannel"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	passed  = "\u2713"
	failed  = "\u2717"
	tmpFile = "/tmp/filechannels.pub"
)

var (
	splitter = []byte("\r\n")

	blueChip = message.PushMessage{
		Sensor:      "BlueChip",
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	ping = message.PushMessage{
		Sensor:      "Ping",
		Category:    "Ping Notificiation",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	crum = message.PushMessage{
		Sensor:      "Crum Stream",
		Category:    "WebRTC Crum Stream",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}
)

// TestFileChannel validates the behaviour of the FileChannel.
func TestFileChannel(t *testing.T) {
	t.Logf("Given the need to sync PushMessages to files")
	{
		t.Logf("When giving a file[%q] and a good configuration with no filters", tmpFile)
		{

			fc := fschannel.New(fschannel.FileConfig{
				MaxSize:         5,
				Timeout:         4 * time.Second,
				DestinationFile: tmpFile,
			})

			fc.Send([]message.PushMessage{blueChip, crum, ping})

			fc.Wait()

			data, err := ioutil.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully read %s file", failed, tmpFile)
			}
			t.Logf("\t%s\t Should have successfully read %s file", passed, tmpFile)

			if contents := bytes.Split(data, splitter); len(contents) == 3 {
				t.Fatalf("\t%s\t Should have successfully match content length in %s to %d", failed, tmpFile, 3)
			}
			t.Logf("\t%s\t Should have successfully match content length in %s to %d", passed, tmpFile, 3)
		}
	}
}
