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

package honeytrap_test

import (
	"bytes"
	"net"
	"os"
	"strings"
	"testing"
)

// Test that we can create a Redis service, connect to it, send INFO and receive a string
// https://redis.io/topics/protocol
func TestRedis(t *testing.T) {
	tmpConf, p := runWithConfig(serviceWithPort("redis", "tcp/6379"))
	defer os.Remove(tmpConf)
	defer p.Kill()
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		t.Error(err)
	}
	_, err = conn.Write([]byte("*1\r\n$4\r\nINFO\r\n"))
	if err != nil {
		t.Error(err)
	}
	resp := make([]byte, 100)
	_, err = conn.Read(resp)
	if err != nil {
		t.Error(err)
	}
	if resp[0] != byte('$') {
		t.Errorf("Expected bulk string ($), got 0x%X", resp[0])
	}
	if !bytes.Contains(resp, []byte("\r\n")) {
		t.Error("No CRLF found in response")
	}
}

// Test that nmap recognizes the Redis service
func TestNmapRedis(t *testing.T) {
	tmpConf, p := runWithConfig(serviceWithPort("redis", "tcp/6379"))
	defer os.Remove(tmpConf)
	defer p.Kill()
	product := nmapIdentify(t, "6379")
	if !strings.Contains(product, "Redis key-value store") {
		t.Errorf("Expected 'Redis key-value store' identification, found '%s'", product)
	}
}
