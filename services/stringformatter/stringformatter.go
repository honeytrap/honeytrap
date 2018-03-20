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

// Package stringformatter formats strings based upon a given template with a time and text element.
package stringformatter

import (
	"strings"
	"text/template"
	"time"
)

type strFormat struct {
	templ *template.Template
}

func New(templ string) (*strFormat, error) {
	t, err := template.New("").Funcs(template.FuncMap{
		"timefmt": func(tm time.Time, fmt string) string {
			if fmt == "" {
				fmt = time.RFC3339
			}

			return tm.Format(fmt)
		},
	}).Parse(templ)
	if err != nil {
		return nil, err
	}

	return &strFormat{
		templ: t,
	}, nil
}

// Format returns the string formatted from a template.
// An empty <time format> will render as RFC3339
func (s *strFormat) Format(tm time.Time, vartext string) string {

	var parsed strings.Builder

	if err := s.templ.Execute(&parsed, struct {
		Time    time.Time
		VarText string
	}{
		Time:    tm,
		VarText: vartext,
	}); err != nil {
		return ""
	}

	return parsed.String()
}
