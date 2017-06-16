package pushers

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/event"
	logging "github.com/op/go-logging"
)

//=================================================================================================

var log = logging.MustGetLogger("honeytrap:channels")

//================================================================================

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send(*event.Event)
}

//=======================================================================================================

// ChannelGenerator defines a function type which returns a Channel created
// from a primitive.
type ChannelGenerator func(toml.MetaData, toml.Primitive) (Channel, error)

// TODO(alex): Decide if we need a mutex to secure things concurrently.
// We assume it will never be read/written to concurrently.
var backends = struct {
	b map[string]ChannelGenerator
}{
	b: make(map[string]ChannelGenerator),
}

// RegisterBackend adds the giving generator to the global generator lists.
func RegisterBackend(name string, generator ChannelGenerator) ChannelGenerator {
	backends.b[name] = generator
	return generator
}

// NewBackend returns a new Channel of the giving name with the provided toml.Primitive.
func NewBackend(name string, meta toml.MetaData, primi toml.Primitive) (Channel, error) {
	log.Debug("Initializing backend : %#q", name)

	maker, ok := backends.b[name]
	if !ok {
		return nil, fmt.Errorf("Backend %q not found", name)
	}

	return maker(meta, primi)
}

//=======================================================================================================

// ChannelsFrom generates a new set of channels from the provided configuration
// adding them as subscribers to a EventBus instance.
func ChannelsFrom(conf *config.Config, bus *EventBus) {
	for _, cb := range conf.Channels {
		if err := MakeFilter(bus, conf, cb); err != nil {
			log.Errorf("Failed creating filter channels : %#q: %#v", cb, err)
			continue
		}
	}
}

// MakeFilter returns a slice of Channels which match the giving criterias
// associated with the provided config.ChannelConfig.
func MakeFilter(bus *EventBus, config *config.Config, conf config.ChannelConfig) error {
	// Generate all filters for the channel's backends
	for _, backend := range conf.Backends {
		// Retrieve backend configuration.
		backendPrimitive, ok := config.Backends[backend]
		if !ok {
			log.Errorf("Application has no backend named %q", backend)
			continue
		}

		var item = struct {
			Backend string `toml:"backend"`
		}{}

		if err := config.PrimitiveDecode(backendPrimitive, &item); err != nil {
			log.Errorf("Could not decode configuration for backend: %q", backend)
			continue
		}

		// Attempt to create backend channel for master with the giving
		// channel's name and config toml.Primitive.
		base, err := NewBackend(item.Backend, config.MetaData, backendPrimitive)
		if err != nil {
			return err
		}

		channel := base

		if config.Token != "" {
			channel = TokenChannel(channel, config.Token)
		}

		if len(conf.Categories) != 0 {
			channel = FilterChannel(channel, RegexFilterFunc("category", conf.Categories))
		}

		if len(conf.Sensors) != 0 {
			channel = FilterChannel(channel, RegexFilterFunc("sensor", conf.Sensors))
		}

		if len(conf.Events) != 0 {
			channel = FilterChannel(channel, RegexFilterFunc("event", conf.Events))
		}

		bus.Subscribe(channel)
	}

	return nil
}

//================================================================================
