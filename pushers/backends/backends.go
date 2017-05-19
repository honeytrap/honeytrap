package backends

import (
	"errors"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/backends/elasticsearch"
	"github.com/honeytrap/honeytrap/pushers/backends/fschannel"
	"github.com/honeytrap/honeytrap/pushers/backends/honeytrap"
	"github.com/honeytrap/honeytrap/pushers/backends/slack"
)

//=======================================================================================================

// SlackBackend defines a function to return a pushers.Channel which delivers
// new messages to a giving underline slack channel defined by the configuration
// retrieved from the giving toml.Primitive.
func SlackBackend(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig slack.APIConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.Host == "" {
		return nil, errors.New("slack.APIConfig Invalid: Host can not be empty")
	}

	if apiconfig.Token == "" {
		return nil, errors.New("slack.APIConfig Invalid: Token can not be empty")
	}

	return slack.New(apiconfig), nil
}

//=======================================================================================================

// HoneytrapBackend defines a function to return a pushers.Channel which delivers
// new messages to a giving underline honeytrap API defined by the configuration
// retrieved from the giving toml.Primitive.
func HoneytrapBackend(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig honeytrap.TrapConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.Host == "" {
		return nil, errors.New("honeytrap.TrapConfig Invalid: Host can not be empty")
	}

	if apiconfig.Token == "" {
		return nil, errors.New("honeytrap.TrapConfig Invalid: Token can not be empty")
	}

	return honeytrap.New(apiconfig), nil
}

//=======================================================================================================

// ElasticBackend defines a function to return a pushers.Channel which delivers
// new messages to a giving underline ElasticSearch API defined by the configuration
// retrieved from the giving toml.Primitive.
func ElasticBackend(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var apiconfig elasticsearch.SearchConfig

	if err := meta.PrimitiveDecode(data, &apiconfig); err != nil {
		return nil, err
	}

	if apiconfig.Host == "" {
		return nil, errors.New("elasticsearch.SearchConfig Invalid: Host can not be empty")
	}

	return elasticsearch.New(apiconfig), nil
}

//=======================================================================================================

var (
	defaultMaxSize  = 1024 * 1024 * 1024
	defaultWaitTime = 5 * time.Second
)

// FileConfProxy is a struct used to received toml decoded values which are string for
// use in generating proper values for a file channel config.
type FileConfProxy struct {
	Timeout     string `toml:"ms"`
	Destination string `toml:"file"`
	MaxSize     int    `toml:"max_size"`
}

// FileBackend defines a function to return a pushers.Channel which delivers
// new messages to a giving underline system file, defined by the configuration
// retrieved from the giving toml.Primitive.
func FileBackend(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var baseconfig FileConfProxy

	if err := meta.PrimitiveDecode(data, &baseconfig); err != nil {
		return nil, err
	}

	if baseconfig.Destination == "" {
		return nil, errors.New("fschannel.FileConfig Invalid: DestinationFile can not be empty")
	}

	var ms time.Duration

	ms = config.MakeDuration(baseconfig.Timeout, int(defaultWaitTime))

	return fschannel.New(fschannel.FileConfig{
		Timeout:         ms,
		MaxSize:         baseconfig.MaxSize,
		DestinationFile: baseconfig.Destination,
	}), nil
}

//=======================================================================================================

func init() {
	pushers.RegisterBackend("file", FileBackend)
	pushers.RegisterBackend("slack", SlackBackend)
	pushers.RegisterBackend("honeytrap", HoneytrapBackend)
	pushers.RegisterBackend("elasticsearch", ElasticBackend)
}
