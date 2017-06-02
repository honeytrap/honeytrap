package fschannel_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/backends/fschannel"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/honeytrap/honeytrap/utils/tests"
)

const (
	passed  = "\u2713"
	failed  = "\u2717"
	tmpFile = "/tmp/filechannels.pub"
)

var (
	splitter = []byte("\r\n")

	blueChip = message.BasicEvent{
		Sensor:      "BlueChip",
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	ping = message.BasicEvent{
		Sensor:      "Ping",
		Category:    "Ping Notificiation",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	crum = message.BasicEvent{
		Sensor:      "Crum Stream",
		Category:    "WebRTC Crum Stream",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}
)

// TestFileBackend validates the behaviour of the FileBackend.
func TestFileBackend(t *testing.T) {
	t.Logf("Given the need to sync PushMessages to files")
	{
		t.Logf("When giving a file[%q] and a good configuration with no filters", tmpFile)
		{

			fc := fschannel.New(fschannel.FileConfig{
				MaxSize: 5,
				Timeout: "2s",
				File:    tmpFile,
			})

			fc.Send(crum)
			fc.Send(ping)
			fc.Send(blueChip)

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

func TestFileGenerator(t *testing.T) {
	tomlConfig := `
	backend = "file"
	file = "/store/files/pushers.pub"
	ms = "50s"
	max_size = 3000`

	var config toml.Primitive

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	var backend = struct {
		Backend string `toml:"backend"`
	}{}

	if err := meta.PrimitiveDecode(config, &backend); err != nil {
		tests.Failed("Should have successfully parsed backend name.")
	}
	tests.Passed("Should have successfully parsed backend name.")

	if backend.Backend != "file" {
		tests.Failed("Should have properly unmarshalled value of config.Backend : %q.", backend.Backend)
	}
	tests.Passed("Should have properly unmarshalled value of config.Backend.")

	if _, err := fschannel.NewWith(meta, config); err != nil {
		tests.Failed("Should have successfully created new  backend:: %+q.", err)
	}
	tests.Passed("Should have successfully created new  backend.")
}
