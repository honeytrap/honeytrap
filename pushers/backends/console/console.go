package console

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/pushers/event"
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

	ch     chan event.Map
	config Config
}

// New returns a new instance of a FileBackend.
func New(c Config) *ConsoleBackend {
	ch := make(chan event.Map, 100)

	backend := ConsoleBackend{
		Writer: os.Stdout,
		ch:     ch,
	}

	go backend.run()

	return &backend
}

// NewWith defines a function to return a pushers.Backend which delivers
// new event.s to a giving underline system file, defined by the configuration
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
			buf := make([]byte, 4)

			n := utf8.EncodeRune(buf, rune)
			o += fmt.Sprintf("\\x%s", hex.EncodeToString(buf[:n]))
			continue
		}

		o += string(rune)
	}

	return o
}

func (b ConsoleBackend) run() {
	for {
		e := <-b.ch

		params := []string{}
		for k, v := range e {
			switch v.(type) {
			case net.IP:
				params = append(params, fmt.Sprintf("%s=%s", k, v.(net.IP).String()))
			case uint32:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case uint16:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case uint8:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case uint:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case int32:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case int16:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case int8:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case int:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case string:
				params = append(params, fmt.Sprintf("%s=%s", k, printify(v.(string))))
			default:
				params = append(params, fmt.Sprintf("%s=%#v", k, v))
			}
		}

		fmt.Fprintf(b.Writer, "%s > %s > %s\n", e["sensor"], e["category"], strings.Join(params, ", "))
	}
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (b *ConsoleBackend) Send(e *event.Event) {
	b.ch <- e.Map()
}
