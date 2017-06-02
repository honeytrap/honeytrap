package honeytrap_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/backends/honeytrap"
	"github.com/honeytrap/honeytrap/utils/tests"
)

func TestHoneytrapGenerator(t *testing.T) {
	tomlConfig := `
	backend = "honeytrap"
	host = "https://hooks.slack.com/services/"
	token = "KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"`

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

	if backend.Backend != "honeytrap" {
		tests.Failed("Should have properly unmarshalled value of config.Backend")
	}
	tests.Passed("Should have properly unmarshalled value of config.Backend.")

	if _, err := honeytrap.NewWith(meta, config); err != nil {
		tests.Failed("Should have successfully created new honeytrap backend: %+q.", err)
	}
	tests.Passed("Should have successfully created new honeytrap backend.")
}
