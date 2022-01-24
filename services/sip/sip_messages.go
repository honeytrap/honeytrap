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
	"fmt"
)

func (s *sipService) OptionsMethod(Request *request) map[string][]string {
	r := Request
	return map[string][]string{
		"Via":            []string{r.SIPVersion + "/TCP " + r.User + ";branch=z9hG4bK" + randomID(20)},
		"From":           []string{"<" + r.UriType + ":" + r.User + "@" + r.RemoteIP + ">;tag=" + randomID(20)},
		"To":             []string{"<sip:nm2@" + r.LocalIP + ">"},
		"Call-ID":        []string{randomID(20) + "@" + r.RemoteIP},
		"CSeq":           []string{"42 " + r.Method},
		"Max-Forwards":   []string{"70"},
		"Contact":        []string{"<" + r.UriType + ":" + r.User + "@" + r.RemoteIP + ">;transport=tcp"},
		"Content-Length": []string{"0"},
		"Accept":         []string{"application/sdp"},
	}
}

func (s *sipService) InviteMethod(Request *request) map[string][]string {
	r := Request
	str := fmt.Sprintf("%v", len(s.InviteBody(r)))
	return map[string][]string{
		"Via":            []string{r.SIPVersion + "/TCP " + r.User + ";branch=z9hG4bK" + randomID(20)},
		"From":           []string{"<" + r.UriType + ":" + r.User + "@" + r.RemoteIP + ">;tag=" + randomID(20)},
		"To":             []string{"<sip:nm2@" + r.LocalIP + ">"},
		"Call-ID":        []string{randomID(20) + "@" + r.RemoteIP},
		"CSeq":           []string{"15 " + r.Method},
		"Max-Forwards":   []string{"70"},
		"Contact":        []string{"<" + r.UriType + ":" + r.User + "@" + r.RemoteIP + ">;transport=tcp"},
		"Content-Length": []string{str},
		"Content-Type":   []string{"application/sdp"},
		"User-Agent":     []string{"Ekiga SIP Softphone"},
	}
}

func (s *sipService) PublishMethod(Request *request) map[string][]string {
	r := Request
	return map[string][]string{
		"Via":            []string{r.SIPVersion + "/TCP " + r.User + ";branch=z9hG4bK" + randomID(20)},
		"From":           []string{"<" + r.UriType + ":" + r.User + "@" + r.RemoteIP + ">;tag=" + randomID(20)},
		"To":             []string{"<sip:nm2@" + r.LocalIP + ">"},
		"Call-ID":        []string{randomID(20) + "@" + r.RemoteIP},
		"CSeq":           []string{"1 " + r.Method},
		"Max-Forwards":   []string{"70"},
		"Expires":        []string{"3600"},
		"Event":          []string{"presence"},
		"Content-Length": []string{"0"},
		"Content-Type":   []string{"application/sdp"},
	}
}

func (s *sipService) InviteBody(Request *request) string {
	r := Request
	return fmt.Sprintf(`
v=0
o=%s:%s@%s 1 16 IN IP4 %s
s=%s:%s@%s
c=IN IP4 %s
t=0 0
m=audio 5000 RTP/AVP 0 8 18 4 120
a=rtpmap:0 PCMU/8000/1
a=rtpmap:8 PCMA/8000/1
a=rtpmap:18 G729/8000/1
a=fmtp:18 annexb=no
a=rtpmap:4 G723/8000/1
a=rtpmap:120 telephone-event/8000/1`, r.UriType, r.User, r.RemoteIP, r.RemoteIP, r.UriType, r.User, r.RemoteIP, r.RemoteIP)
}
