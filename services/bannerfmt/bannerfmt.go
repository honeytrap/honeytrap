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
package bannerfmt

import (
	"strings"
	"text/template"
	"time"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("bannerformatter")

type BannerFmt struct {
	templ *template.Template

	data interface{}
}

// New creates a new template using 'templ' as format
// arguments:
// templ - go template string, see https://golang.org/pkg/text/template/
// data  - the data structure to use in your template
//
// template functions:
// now [time-format string] - the current time formatted as `time-format`
// timefmt [tm time.Time] [time-format string] - the time `tm` formatted as `time-format`
// time-format:
//   example time format eg. `2018-01-20 15:00`
func New(templ string, data interface{}) (*BannerFmt, error) {

	t, err := template.New("").Funcs(template.FuncMap{
		"timefmt": func(tm time.Time, fmt string) string {
			if fmt == "" {
				fmt = time.RFC3339
			}
			return tm.Format(fmt)
		},
		"now": func(fmt string) string {
			if fmt == "" {
				return time.Now().String()
			}
			return time.Now().Format(fmt)
		},
	}).Parse(templ)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	return &BannerFmt{
		templ: t,
		data:  data,
	}, nil
}

// String returns the formatted banner string
// On error it returns an empty or partially formatted string.
func (b *BannerFmt) String() string {
	var parsed strings.Builder

	if err := b.templ.Execute(&parsed, b.data); err != nil {
		log.Debug(err.Error())
	}

	return parsed.String()
}

// Set replaces the data where the banner is rendered from
//   this should be of the same type
func (b *BannerFmt) Set(data interface{}) {
	b.data = data
}
