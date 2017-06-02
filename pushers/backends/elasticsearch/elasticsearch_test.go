package elasticsearch_test

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers/backends/elasticsearch"
	"github.com/honeytrap/honeytrap/utils/tests"
)

func TestElasticSearchGenerator(t *testing.T) {
	tomlConfig := `
	backend = "elasticsearch"
	host = "https://api.elastic.com/db/"`

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

	if backend.Backend != "elasticsearch" {
		tests.Failed("Should have properly unmarshalled value of config.Backend.")
	}
	tests.Passed("Should have properly unmarshalled value of config.Backend.")

	if _, err := elasticsearch.NewWith(meta, config); err != nil {
		tests.Failed("Should have successfully created new elasticsearch backend:: %+q.", err)
	}
	tests.Passed("Should have successfully created new elasticsearch backend.")
}
