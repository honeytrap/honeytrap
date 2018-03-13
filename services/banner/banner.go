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
package banner

import (
	"strings"
	"text/template"
	"time"
)

// Create a Banner template "[Host] [Text] [datetime]"
func New(Host, Text string, DateTime bool) (*bannerT, error) {
	var tmpl strings.Builder

	if Host != "" {
		tmpl.WriteString(Host)
		tmpl.WriteString(" ")
	}

	if Text != "" {
		tmpl.WriteString(Text)
		tmpl.WriteString(" ")
	}

	if DateTime {
		tmpl.WriteString("{{datetime}}")
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"datetime": func() string {
			return time.Now().Format(time.RFC3339)
		},
	}).Parse(tmpl.String())
	if err != nil {
		return nil, err
	}

	return &bannerT{
		tmpl: t,
	}, nil
}

type bannerT struct {
	tmpl *template.Template
}

// Banner string with updated datetime (if set to true in New)
func (t *bannerT) Banner() string {
	var parsed strings.Builder

	if err := t.tmpl.Execute(&parsed, ""); err != nil {
		return ""
	}

	return parsed.String()
}
