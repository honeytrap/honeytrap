package fschannel_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/fschannel"
)

const (
	passed  = "\u2713"
	failed  = "\u2717"
	tmpFile = "/tmp/filechannels.pub"
)

var (
	splitter = []byte("\r\n")

	blueChip = &pushers.PushMessage{
		Sensor:      "BlueChip",
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	ping = &pushers.PushMessage{
		Sensor:      "Ping",
		Category:    "Ping Notificiation",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	crum = &pushers.PushMessage{
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
		fc := fschannel.New()

		t.Logf("When giving a file[%q] and a bad configuration", tmpFile)
		{

			err := fc.UnmarshalConfig(map[string]interface{}{
				"target": tmpFile,
				"filters": map[string]interface{}{
					"sensor": "^ping",
				},
			})

			if err == nil {
				t.Fatalf("\t%s\t Should have successfully failed to parse configuration", failed)
			}
			t.Logf("\t%s\t Should have successfully failed to parse configuration", passed)
		}

		t.Logf("When giving a file[%q] and a good configuration", tmpFile)
		{

			err := fc.UnmarshalConfig(map[string]interface{}{
				"ms":   "20s",
				"size": "5",
				"file": tmpFile,
				"filters": map[string]interface{}{
					"sensor": "[^ping]",
				},
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully parsed configuration", failed)
			}
			t.Logf("\t%s\t Should have successfully parsed configuration", passed)

			fc.Send([]*pushers.PushMessage{blueChip, crum, ping})

			fc.Wait()
		}

		t.Logf("When giving a file[%q] and a good configuration with no filters", tmpFile)
		{

			err := fc.UnmarshalConfig(map[string]interface{}{
				"ms":   "20s",
				"size": "5",
				"file": tmpFile,
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully parsed configuration", failed)
			}
			t.Logf("\t%s\t Should have successfully parsed configuration", passed)

			fc.Send([]*pushers.PushMessage{blueChip, crum, ping})

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

		t.Logf("When giving a file[%q] and a good configuration with filters", tmpFile)
		{

			err := fc.UnmarshalConfig(map[string]interface{}{
				"ms":   "20s",
				"size": "5",
				"file": tmpFile,
				"filters": map[string]interface{}{
					"sensor": "[^Ping]",
				},
			})

			if err != nil {
				t.Fatalf("\t%s\t Should have successfully parsed configuration", failed)
			}
			t.Logf("\t%s\t Should have successfully parsed configuration", passed)

			fc.Send([]*pushers.PushMessage{blueChip, crum, ping})

			fc.Wait()

			data, err := ioutil.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("\t%s\t Should have successfully read %s file", failed, tmpFile)
			}
			t.Logf("\t%s\t Should have successfully read %s file", passed, tmpFile)

			if contents := bytes.Split(data, splitter); len(contents) == 2 {
				t.Fatalf("\t%s\t Should have successfully match content length in %s to %d", failed, tmpFile, 2)
			}
			t.Logf("\t%s\t Should have successfully match content length in %s to %d", passed, tmpFile, 2)
		}
	}
}
