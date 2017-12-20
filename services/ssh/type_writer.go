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
package ssh

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

func NewTypeWriterReadCloser(r io.ReadWriteCloser) *TypeWriterReadWriteCloser {
	return &TypeWriterReadWriteCloser{ReadWriteCloser: r, time: time.Now()}
}

type TypeWriterReadWriteCloser struct {
	io.ReadWriteCloser

	time   time.Time
	buffer bytes.Buffer
}

func sanitize(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "<br/>", -1)
	s = strings.Replace(s, "'", "\\'", -1)
	s = strings.Replace(s, "\b", "<backspace>", -1)
	return s
}

func (lr *TypeWriterReadWriteCloser) Write(p []byte) (n int, err error) {
	return lr.ReadWriteCloser.Write(p)
}

func (lr *TypeWriterReadWriteCloser) Read(p []byte) (n int, err error) {
	n, err = lr.ReadWriteCloser.Read(p)

	now := time.Now()
	lr.buffer.WriteString(fmt.Sprintf(".wait(%d)", int(now.Sub(lr.time).Seconds()*1000)))
	lr.buffer.WriteString(fmt.Sprintf(".put('%s')", sanitize(string(p[:n]))))
	lr.time = now

	log.Debugf(sanitize(string(p[:n])))
	return n, err
}

func (lr *TypeWriterReadWriteCloser) String() string {
	return lr.buffer.String()
}

func (lr *TypeWriterReadWriteCloser) Close() error {
	return lr.ReadWriteCloser.Close()
}
