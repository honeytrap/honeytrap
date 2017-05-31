package console

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var (
	_ = pushers.RegisterBackend("console", NewWith)
)

var (
	log = logging.MustGetLogger("console")
)

// Config defines the config used to setup the ConsoleBackend.
type Config struct {
}

// ConsoleBackend provides a backend for outputing event details directly to
// the current console.
type ConsoleBackend struct {
	io.Writer

	config Config
}

// New returns a new instance of a FileBackend.
func New(c Config) *ConsoleBackend {
	return &ConsoleBackend{
		Writer: os.Stdout,
	}
}

// NewWith defines a function to return a pushers.Backend which delivers
// new messages to a giving underline system file, defined by the configuration
// retrieved from the giving toml.Primitive.
func NewWith(meta toml.MetaData, data toml.Primitive) (pushers.Channel, error) {
	var config Config

	if err := meta.PrimitiveDecode(data, &config); err != nil {
		return nil, err
	}

	return New(config), nil
}

func printify(s string) string {
	o := ""
	for _, rune := range s {
		if !unicode.IsPrint(rune) {
			o += "\xa4"
			continue
		}

		o += string(rune)
	}

	return o
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (f *ConsoleBackend) Send(message message.Event) {
	params := []string{}
	for k, v := range message.Fields() {
		switch v.(type) {
		case string:
			params = append(params, fmt.Sprintf("%s=%s", k, printify(v.(string))))
		default:
			params = append(params, fmt.Sprintf("%s=%#v", k, v))
		}
	}

	category, _, sensor := message.Identity()
	fmt.Fprintf(f.Writer, "%s > %s > %s\n", sensor, category, strings.Join(params, ", "))
}
