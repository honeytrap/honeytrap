package pushers

import (
	"errors"
	"fmt"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	logging "github.com/op/go-logging"
)

//=================================================================================================

var log = logging.MustGetLogger("honeytrap:channels")

//================================================================================

// FilteringChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criterias.
type FilteringChannel struct {
	config   *config.Config
	backends []Channel
	filters  []Filters
}

// NewFilteringChannel returns a new instance of the FilteringChannel.
func NewFilteringChannel(config *config.Config, filters ...Filters) *FilteringChannel {
	var mc FilteringChannel
	mc.config = config
	mc.filters = filters

	return &mc
}

// UnmarshalConfig attempts to unmarshal the provided value into the target
// FilteringChannel.
func (mc *FilteringChannel) UnmarshalConfig(m interface{}) error {
	conf, ok := m.(config.ChannelConfig)
	if !ok {
		return errors.New("Expected to receive a ChannelConfig type")
	}

	// Generate all filters for the channel's backends
	for _, backend := range conf.Backends {

		// Retrieve backend configuration.
		backendPrimitive, ok := mc.config.Backends[backend]
		if !ok {
			return fmt.Errorf("Application has no backend named %q", backend)
		}

		var item = struct {
			Backend string `toml:"backend"`
		}{}

		if err := mc.config.PrimitiveDecode(backendPrimitive, &item); err != nil {
			return err
		}

		// Attempt to create backend channel for master with the giving
		// channel's name and config toml.Primitive.
		newBackend, err := NewBackend(item.Backend, mc.config.MetaData, backendPrimitive)
		if err != nil {
			return err
		}

		mc.backends = append(mc.backends, newBackend)
	}

	mc.filters = append(mc.filters, NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Categories...)...))
	mc.filters = append(mc.filters, NewRegExpFilter(SensorFilterFunc, MakeMatchers(conf.Sensors...)...))
	mc.filters = append(mc.filters, NewRegExpFilter(EventFilterFunc, MakeMatchers(conf.Events...)...))

	return nil
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc *FilteringChannel) Send(msgs ...message.Event) {

	// filter messages with all filters.
	for _, filter := range mc.filters {
		msgs = filter.Filter(msgs...)
	}

	// If no message passes our filtering conditions then we have
	// nothing else to do here.
	if len(msgs) == 0 {
		return
	}

	for _, backend := range mc.backends {
		backend.Send(msgs...)
	}
}
