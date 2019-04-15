// +build linux

// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
