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
