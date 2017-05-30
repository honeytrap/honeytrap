package pushers

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	logging "github.com/op/go-logging"
)

//=================================================================================================

var log = logging.MustGetLogger("honeytrap:channels")

//================================================================================

// Channel defines a interface which exposes a single method for delivering
// PushMessages to a giving underline service.
type Channel interface {
	Send(message.Event)
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
	log.Info("honeytrap.Pusher : Creating %q Backend : %#q", name, primi)

	maker, ok := backends.b[name]
	if !ok {
		return nil, fmt.Errorf("Backend %q maker not found", name)
	}

	return maker(meta, primi)
}

//=======================================================================================================

// ChannelsFrom generates a new set of channels from the provided configuration
// adding them as subscribers to a EventBus instance.
func ChannelsFrom(conf *config.Config, bus *EventBus) {
	for _, cb := range conf.Channels {
		if err := MakeFilter(bus, conf, cb); err != nil {
			log.Errorf("Failed creating filter channels : %#q", cb)
			continue
		}
	}
}

//================================================================================

type filterChannel struct {
	Channel

	Filter FilterGroup
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc filterChannel) Send(msgs message.Event) {
	for _, item := range mc.Filter.Filter(msgs) {
		mc.Channel.Send(item)
	}
}

// FilterChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criterias.
func FilterChannel(channel Channel, filter FilterGroup) Channel {
	return filterChannel{
		Channel: channel,
		Filter:  filter,
	}
}

//================================================================================

// MakeFilter returns a slice of Channels which match the giving criterias
// associated with the provided config.ChannelConfig.
func MakeFilter(bus *EventBus, config *config.Config, conf config.ChannelConfig) error {
	var filters FilterGroup

	filters.Add(NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Categories...)...))
	filters.Add(NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Sensors...)...))
	filters.Add(NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Events...)...))

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

		bus.Subscribe(FilterChannel(base, filters))
	}

	return nil
}

//================================================================================
