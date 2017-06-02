package pushers_test

import (
	"testing"

	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
)

const (
	passed = "\u2713"
	failed = "\u2717"
)

var (
	blueChip = message.BasicEvent{
		Sensor:      "BlueChip",
		Category:    "Chip Integrated",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	ping = message.BasicEvent{
		Sensor:      "Ping",
		Category:    "Ping Notificiation",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}

	crum = message.BasicEvent{
		Sensor:      "Crum Stream",
		Category:    "WebRTC Crum Stream",
		SessionID:   "4334334-3433434-34343-FUD",
		ContainerID: "56454-5454UDF-2232UI-34FGHU",
		Data:        "Hello World!",
	}
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
