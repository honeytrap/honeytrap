package stdout

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
	"github.com/honeytrap/honeytrap/pushers/message"
	"github.com/op/go-logging"
)

var (
	_ = pushers.RegisterBackend("stdout", NewWith)
)

var (
	log = logging.MustGetLogger("stdout")
)

// Config defines the config used to setup the StdoutBackend.
type Config struct {
}

// FileBackend defines a struct which implements the pushers.Pusher interface
// and allows us to write PushMessage updates into a giving file path. Mainly for
// the need to sync PushMessage to local files for persistence.
// File paths provided are either created with a append mode if they already
// exists else will be created. FileBackend will also restrict filesize to a max of 1gb by default else if
// there exists a max size set in configuration, then that will be used instead,
// also the old file will be renamed with the current timestamp and a new file created.
type StdoutBackend struct {
	io.Writer

	config Config
}

// New returns a new instance of a FileBackend.
func New(c Config) *StdoutBackend {
	return &StdoutBackend{
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
			buf := make([]byte, 4)

			n := utf8.EncodeRune(buf, rune)
			o += fmt.Sprintf("\\x%s", hex.EncodeToString(buf[:n]))
			continue
		}

		o += string(rune)
	}

	return o
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (f *StdoutBackend) Send(message message.Event) {
	params := []string{}
	for k, v := range message.Details {
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

	fmt.Fprintf(f.Writer, "%s > %s > %s\n", message.Sensor, message.Category, strings.Join(params, ", "))
}
