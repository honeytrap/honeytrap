/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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

// Package bannerfmt formats strings based upon a given template with a time and text element.
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
