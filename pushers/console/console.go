/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
package console

import (
	"encoding/hex"
	"fmt"
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
