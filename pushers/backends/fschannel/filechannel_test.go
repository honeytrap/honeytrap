package fschannel_test

import (
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/backends/fschannel"
	"github.com/honeytrap/honeytrap/pushers/event"
	"github.com/honeytrap/honeytrap/utils/tests"
)

const (
	passed  = "\u2713"
	failed  = "\u2717"
	tmpFile = "/tmp/filechannels.pub"
)

var (
	blueChip = event.New(
		event.Sensor("BlueChip"),
		event.Category("Chip Integrated"),
	)

	ping = event.New(
		event.Sensor("Ping"),
		event.Category("Ping Notification"),
	)

	crum = event.New(
		event.Sensor("Crum Stream"),
		event.Category("WebRTC Crum Stream"),
	)
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

			if len(data) == 0 {
				t.Fatalf("\t%s\t Should have successfully have file size greater than 0", failed)
			}
			t.Logf("\t%s\t Should have successfully have file size greater than 0", passed)
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
