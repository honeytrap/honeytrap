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
package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
)

// succeedMark is the Unicode codepoint for a check mark.
const succeedMark = "\u2713"

var logger = log.New(os.Stdout, "", log.Lshortfile)

// Info logs the info message using the giving message and values.
func Info(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t-\t %s\n", fmt.Sprintf(message, val...)))
	}
}

// Passed logs the failure message using the giving message and values.
func Passed(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", succeedMark, fmt.Sprintf(message, val...)))
	}
}

// failedMark is the Unicode codepoint for an X mark.
const failedMark = "\u2717"

// Failed logs the failure message using the giving message and values.
func Failed(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", failedMark, fmt.Sprintf(message, val...)))
	}

	os.Exit(1)
}

// Errored logs the error message using the giving message and values.
func Errored(message string, val ...interface{}) {
	if testing.Verbose() {
		logger.Output(2, fmt.Sprintf("\t%s\t %s\n", failedMark, fmt.Sprintf(message, val...)))
	}
}
