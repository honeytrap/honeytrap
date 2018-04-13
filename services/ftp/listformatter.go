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
package ftp

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type listFormatter []os.FileInfo

// Short returns a string that lists the collection of files by name only,
// one per line
func (formatter listFormatter) Short() []byte {
	var buf bytes.Buffer
	for _, file := range formatter {
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.Bytes()
}

// Detailed returns a string that lists the collection of files with extra
// detail, one per line
func (formatter listFormatter) Detailed() []byte {
	var buf bytes.Buffer
	for _, file := range formatter {
		fmt.Fprintf(&buf, file.Mode().String())
		//fmt.Fprintf(&buf, " 1 %s %s ", file.Owner(), file.Group())
		fmt.Fprintf(&buf, lpad(strconv.Itoa(int(file.Size())), 12))
		fmt.Fprintf(&buf, file.ModTime().Format(" Jan _2 15:04 "))
		fmt.Fprintf(&buf, "%s\r\n", file.Name())
	}
	fmt.Fprintf(&buf, "\r\n")
	return buf.Bytes()
}

func lpad(input string, length int) (result string) {
	if len(input) < length {
		result = strings.Repeat(" ", length-len(input)) + input
	} else if len(input) == length {
		result = input
	} else {
		result = input[0:length]
	}
	return
}
