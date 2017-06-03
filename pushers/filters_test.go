package pushers_test

import (
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/event"
)

const (
	passed = "\u2713"
	failed = "\u2717"
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

func TestRegExpFilter(t *testing.T) {
	t.Logf("Given the need to filter giving messages based on message fields")
	{

		t.Logf("\tWhen filtering is based on the 'sensor' field")
		{

			filter := pushers.NewRegExpFilter(pushers.SensorFilterFunc, pushers.MakeMatchers("Pingo")...)

			if dl := len(filter.Filter(blueChip, ping, crum)); dl != 0 {
				t.Fatalf("\t%s\t Should have successfully filtered out all messages: %d.", failed, dl)
			}
			t.Logf("\t%s\t Should have successfully filtered out all messages.", passed)

		}

		t.Logf("\tWhen filtering is based on the 'category' field")
		{

			filter := pushers.NewRegExpFilter(pushers.CategoryFilterFunc, pushers.MakeMatchers("^WebRTC")...)

			if dl := len(filter.Filter(blueChip, ping, crum)); dl != 1 {
				t.Fatalf("\t%s\t Should have successfully filtered all but one message: %d.", failed, dl)
			}
			t.Logf("\t%s\t Should have successfully filtered all but one message.", passed)

		}

	}
}
