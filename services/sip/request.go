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

package sip

import (
	"strings"
)

func (s *sipService) parseURI() bool {
	if !strings.HasPrefix(s.Uri, "sip:") {
		return false
	}

	request := strings.Split(s.Uri, ":")

	if strings.Contains(request[1], "@") {
		result := strings.Split(request[1], "@")
		s.Username = result[0]
		s.Domain = result[1]
	} else {
		s.Username = request[1]
		s.Domain = request[1]
	}

	return true
}

func (s *sipService) checkRequest(line string) bool {
	ok := s.parseURI()
	if !ok {
		return ok
	}

	for i, _ := range Map_Method {
		if s.Method == Map_Method[i] {
			ok = true
			break
		}
	}

	if !ok {
		return ok
	}

	if s.SIPVersion != "SIP/2.0" {
		return false
	}

	return true
}
