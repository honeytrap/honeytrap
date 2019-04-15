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
package ipv4

import (
	"encoding/binary"
	"errors"
	"net"
	"unsafe"
)

var (
	errMissingAddress           = errors.New("missing address")
	errMissingHeader            = errors.New("missing header")
	errHeaderTooShort           = errors.New("header too short")
	errBufferTooShort           = errors.New("buffer too short")
	errInvalidConnType          = errors.New("invalid conn type")
	errOpNoSupport              = errors.New("operation not supported")
	errNoSuchInterface          = errors.New("no such interface")
	errNoSuchMulticastInterface = errors.New("no such multicast interface")

	// See http://www.freebsd.org/doc/en/books/porters-handbook/freebsd-versions.html.
	freebsdVersion uint32

	nativeEndian binary.ByteOrder
)

func init() {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		nativeEndian = binary.LittleEndian
	} else {
		nativeEndian = binary.BigEndian
	}
}

func boolint(b bool) int {
	if b {
		return 1
	}
	return 0
}

func netAddrToIP4(a net.Addr) net.IP {
	switch v := a.(type) {
	case *net.UDPAddr:
		if ip := v.IP.To4(); ip != nil {
			return ip
		}
	case *net.IPAddr:
		if ip := v.IP.To4(); ip != nil {
			return ip
		}
	}
	return nil
}
