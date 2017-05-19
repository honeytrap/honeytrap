package backends_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/backends"
	"github.com/honeytrap/honeytrap/utils/tests"
)

func TestSlackGenerator(t *testing.T) {
	tomlConfig := `
	server = "slack"
	[config]
		host = "https://hooks.slack.com/services/"
		token = "KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"`

	var config config.BackendConfig

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q.", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	if config.Server != "slack" {
		tests.Failed("Should have properly unmarshalled value of config.Server.")
	}
	tests.Passed("Should have properly unmarshalled value of config.Server.")

	if _, err := backends.SlackBackend(meta, config.Config); err != nil {
		tests.Failed("Should have successfully created new slack backend: %+q.", err)
	}
	tests.Passed("Should have successfully created new slack backend.")
}

func TestElasticSearchGenerator(t *testing.T) {
	tomlConfig := `
	server = "elasticsearch"
	[config]
		host = "https://api.elastic.com/db/"`

	var config config.BackendConfig

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q.", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	if config.Server != "elasticsearch" {
		tests.Failed("Should have properly unmarshalled value of config.Server.")
	}
	tests.Passed("Should have properly unmarshalled value of config.Server.")

	if _, err := backends.ElasticBackend(meta, config.Config); err != nil {
		tests.Failed("Should have successfully created new elasticsearch backend:: %+q.", err)
	}
	tests.Passed("Should have successfully created new elasticsearch backend.")
}

func TestHoneytrapGenerator(t *testing.T) {
	tomlConfig := `
	server = "honeytrap"
	[config]
		host = "https://hooks.slack.com/services/"
		token = "KUL6M39MCM/YU16GBD/VOOW9HG60eDfoFBiMF"`

	var config config.BackendConfig

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q.", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	if config.Server != "honeytrap" {
		tests.Failed("Should have properly unmarshalled value of config.Server")
	}
	tests.Passed("Should have properly unmarshalled value of config.Server.")

	if _, err := backends.HoneytrapBackend(meta, config.Config); err != nil {
		tests.Failed("Should have successfully created new honeytrap backend: %+q.", err)
	}
	tests.Passed("Should have successfully created new honeytrap backend.")
}

func TestFiletrapGenerator(t *testing.T) {
	tomlConfig := `
	server = "file"
	[config]
        ms = "50s"
        max_size = 3000
        file = "/store/files/pushers.pub"`

	var config config.BackendConfig

	meta, err := toml.Decode(tomlConfig, &config)
	if err != nil {
		tests.Failed("Should have successfully parsed toml config: %+q", err)
	}
	tests.Passed("Should have successfully parsed toml config.")

	if config.Server != "file" {
		tests.Failed("Should have properly unmarshalled value of config.Server.")
	}
	tests.Passed("Should have properly unmarshalled value of config.Server.")

	if _, err := backends.FileBackend(meta, config.Config); err != nil {
		tests.Failed("Should have successfully created new file backend: %+q", err)
	}
	tests.Passed("Should have successfully created new file backend.")
}
