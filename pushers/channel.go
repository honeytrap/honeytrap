package pushers

import (
	"fmt"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	logging "github.com/op/go-logging"
)

//=================================================================================================

var log = logging.MustGetLogger("honeytrap:channels")

//================================================================================

// FilterChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criterias.
type FilterChannel struct {
	Backend Channel
	Filter  FilterGroup
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc FilterChannel) Send(msgs ...message.Event) {
	mc.Backend.Send(mc.Filter.Filter(msgs...)...)
}

// MakeFilter returns a slice of Channels which match the giving criterias
// associated with the provided config.ChannelConfig.
func MakeFilter(config *config.Config, conf config.ChannelConfig) ([]Channel, error) {
	var filters FilterGroup
	filters = append(filters, NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Categories...)...))
	filters = append(filters, NewRegExpFilter(SensorFilterFunc, MakeMatchers(conf.Sensors...)...))
	filters = append(filters, NewRegExpFilter(EventFilterFunc, MakeMatchers(conf.Events...)...))

	var channels []Channel

	// Generate all filters for the channel's backends
	for _, backend := range conf.Backends {

		// Retrieve backend configuration.
		backendPrimitive, ok := config.Backends[backend]
		if !ok {
			return nil, fmt.Errorf("Application has no backend named %q", backend)
		}

		var item = struct {
			Backend string `toml:"backend"`
		}{}

		if err := config.PrimitiveDecode(backendPrimitive, &item); err != nil {
			return nil, err
		}

		// Attempt to create backend channel for master with the giving
		// channel's name and config toml.Primitive.
		base, err := NewBackend(item.Backend, config.MetaData, backendPrimitive)
		if err != nil {
			return nil, err
		}

		channels = append(channels, FilterChannel{
			Backend: base,
			Filter:  filters,
		})
	}

	return channels, nil
}
