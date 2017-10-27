// +build linux

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
package canary

import (
	"bufio"
	"net"
	"os"
)

// ARPCache defines a slice of ARPEntrys.
type ARPCache []ARPEntry

// Get retrieves the ARPEntry associated with the giving ip.
func (ac ARPCache) Get(ip net.IP) *ARPEntry {
	for _, a := range ac {
		if !a.IP.Equal(ip) {
			continue
		}

		return &a
	}

	return nil
}

// ARPEntry defines a type for containing address and interface detail.
type ARPEntry struct {
	IP              net.IP
	HardwareAddress net.HardwareAddr
	Interface       string
}

func parseARPCache(path string) (ARPCache, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	entries := []ARPEntry{}

	r := bufio.NewReader(f)

	// skip first line
	r.ReadLine()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		parts := splitAtBytes(text, " \r\t\n")
		if len(parts) < 6 {
			continue
		}

		ip := net.ParseIP(parts[0])
		hwaddress, _ := net.ParseMAC(parts[3])

		entries = append(entries, ARPEntry{
			Interface:       parts[5],
			IP:              ip,
			HardwareAddress: hwaddress,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ARPCache(entries), nil
}
