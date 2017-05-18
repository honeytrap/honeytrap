package pushers

import (
	"errors"

	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/pushers/message"
	logging "github.com/op/go-logging"
)

//=================================================================================================

var log = logging.MustGetLogger("honeytrap:channels")

//================================================================================

// ChannelStream defines a type for a slice of Channels implementing objects.
type ChannelStream []Channel

// Send delivers the provided PushMessages to all underline set of Channel implementing
// objects.
func (channels ChannelStream) Send(msg []message.PushMessage) {
	for _, channel := range channels {
		channel.Send(msg)
	}
}

//=================================================================================================

// BackendRegistry defines an interface which prvides a registery of backend Channel
// retrievable through a string key.
type BackendRegistry interface {
	GetBackend(string) (Channel, error)
}

// MasterChannel defines a struct which handles the delivery of giving
// messages to a specific sets of backend channels based on specific criterias.
type MasterChannel struct {
	backends []Channel
	filters  []Filters
	registry BackendRegistry
}

// NewMasterChannel returns a new instance of the MasterChannel.
func NewMasterChannel(br BackendRegistry, filters ...Filters) *MasterChannel {
	var mc MasterChannel
	mc.registry = br
	mc.filters = filters

	return &mc
}

// UnmarshalConfig attempts to unmarshal the provided value into the target
// MasterChannel.
func (mc *MasterChannel) UnmarshalConfig(m interface{}) error {
	conf, ok := m.(config.ChannelConfig)
	if !ok {
		return errors.New("Expected to receive a ChannelConfig type")
	}

	// Generate all filters for the channel's backends
	for _, backend := range conf.Backends {
		bl, err := mc.registry.GetBackend(backend)
		if err != nil {
			return err
		}

		mc.backends = append(mc.backends, bl)
	}

	mc.filters = append(mc.filters, NewRegExpFilter(CategoryFilterFunc, MakeMatchers(conf.Categories...)...))
	mc.filters = append(mc.filters, NewRegExpFilter(SensorFilterFunc, MakeMatchers(conf.Sensors...)...))
	mc.filters = append(mc.filters, NewRegExpFilter(EventFilterFunc, MakeMatchers(conf.Events...)...))

	return nil
}

// Send delivers the slice of PushMessages and using the internal filters
// to filter out the desired messages allowed for all registered backends.
func (mc *MasterChannel) Send(msgs []message.PushMessage) {

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
		backend.Send(msgs)
	}
}
