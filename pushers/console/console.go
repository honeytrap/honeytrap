// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package console

import (
	"encoding/hex"
	"fmt"
	logging "github.com/op/go-logging"
	"io"
	"net"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"

)

var (
	_ = pushers.Register("console", New)
)

var log = logging.MustGetLogger("channels/console")

// Config defines the config used to setup the Console.
type Config struct {
}

// New returns a new instance of a FileBackend.
func New(options ...func(pushers.Channel) error) (pushers.Channel, error) {
	ch := make(chan map[string]interface{}, 100)

	c := Console{
		Writer: os.Stdout,
		ch:     ch,
	}

	for _, optionFn := range options {
		optionFn(&c)
	}

	go c.run()

	return &c, nil
}

// Console provides a backend for outputing event details directly to
// the current console.
type Console struct {
	io.Writer

	ch     chan map[string]interface{}
	config Config
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

func (b Console) run() {
	for e := range b.ch {
		var params []string
		for k, v := range e {
			switch x := v.(type) {
			case net.IP:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case uint32, uint16, uint8, uint,
				int32, int16, int8, int:
				params = append(params, fmt.Sprintf("%s=%d", k, v))
			case time.Time:
				params = append(params, fmt.Sprintf("%s=%s", k, x.String()))
			case string:
				params = append(params, fmt.Sprintf("%s=%s", k, printify(x)))
			default:
				params = append(params, fmt.Sprintf("%s=%#v", k, v))
			}
		}
		sort.Strings(params)
		fmt.Fprintf(b.Writer, "%s > %s > %s\n", e["sensor"], e["category"], strings.Join(params, ", "))
	}
}

// Send delivers the giving if it passes all filtering criteria into the
// FileBackend write queue.
func (b *Console) Send(e event.Event) {
	mp := make(map[string]interface{})

	e.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			mp[keyName] = value
		}
		return true
	})

	b.ch <- mp
}
