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
	"encoding/hex"
	"net"
	"os"
)

// RouteTable defines a slice of Route type.
type RouteTable []Route

// Route defines a Route element detailing given address for a gatewau connection.
type Route struct {
	Interface string

	Gateway     net.IP
	Destination net.IPNet
}

func parseRouteTable(path string) (RouteTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	routes := []Route{}

	r := bufio.NewReader(f)

	// skip first line
	r.ReadLine()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		parts := splitAtBytes(text, " :\r\t\n")
		if len(parts) < 11 {
			continue
		}

		destination, _ := hex.DecodeString(parts[1])
		mask, _ := hex.DecodeString(parts[7])

		gateway, _ := hex.DecodeString(parts[2])

		routes = append(routes, Route{
			Interface: parts[0],
			Gateway:   net.IPv4(gateway[3], gateway[2], gateway[1], gateway[0]),
			Destination: net.IPNet{
				IP:   net.IPv4(destination[0], destination[1], destination[2], destination[3]),
				Mask: net.IPv4Mask(mask[0], mask[1], mask[2], mask[3]),
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return RouteTable(routes), nil
}
