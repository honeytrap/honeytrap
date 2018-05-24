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
package icmp

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

const (
	ICMPv4TypeEchoReply              = 0
	ICMPv4TypeDestinationUnreachable = 3
	ICMPv4TypeSourceQuench           = 4
	ICMPv4TypeRedirect               = 5
	ICMPv4TypeEchoRequest            = 8
	ICMPv4TypeRouterAdvertisement    = 9
	ICMPv4TypeRouterSolicitation     = 10
	ICMPv4TypeTimeExceeded           = 11
	ICMPv4TypeParameterProblem       = 12
	ICMPv4TypeTimestampRequest       = 13
	ICMPv4TypeTimestampReply         = 14
	ICMPv4TypeInfoRequest            = 15
	ICMPv4TypeInfoReply              = 16
	ICMPv4TypeAddressMaskRequest     = 17
	ICMPv4TypeAddressMaskReply       = 18
)

const (
	// DestinationUnreachable
	ICMPv4CodeNet                 = 0
	ICMPv4CodeHost                = 1
	ICMPv4CodeProtocol            = 2
	ICMPv4CodePort                = 3
	ICMPv4CodeFragmentationNeeded = 4
	ICMPv4CodeSourceRoutingFailed = 5
	ICMPv4CodeNetUnknown          = 6
	ICMPv4CodeHostUnknown         = 7
	ICMPv4CodeSourceIsolated      = 8
	ICMPv4CodeNetAdminProhibited  = 9
	ICMPv4CodeHostAdminProhibited = 10
	ICMPv4CodeNetTOS              = 11
	ICMPv4CodeHostTOS             = 12
	ICMPv4CodeCommAdminProhibited = 13
	ICMPv4CodeHostPrecedence      = 14
	ICMPv4CodePrecedenceCutoff    = 15

	// TimeExceeded
	ICMPv4CodeTTLExceeded                    = 0
	ICMPv4CodeFragmentReassemblyTimeExceeded = 1

	// ParameterProblem
	ICMPv4CodePointerIndicatesError = 0
	ICMPv4CodeMissingOption         = 1
	ICMPv4CodeBadLength             = 2

	// Redirect
	// ICMPv4CodeNet  = same as for DestinationUnreachable
	// ICMPv4CodeHost = same as for DestinationUnreachable
	ICMPv4CodeTOSNet  = 2
	ICMPv4CodeTOSHost = 3
)

type icmpv4TypeCodeInfoStruct struct {
	typeStr string
	codeStr *map[uint8]string
}

var (
	icmpv4TypeCodeInfo = map[uint8]icmpv4TypeCodeInfoStruct{
		ICMPv4TypeDestinationUnreachable: {
			"DestinationUnreachable", &map[uint8]string{
				ICMPv4CodeNet:                 "Net",
				ICMPv4CodeHost:                "Host",
				ICMPv4CodeProtocol:            "Protocol",
				ICMPv4CodePort:                "Port",
				ICMPv4CodeFragmentationNeeded: "FragmentationNeeded",
				ICMPv4CodeSourceRoutingFailed: "SourceRoutingFailed",
				ICMPv4CodeNetUnknown:          "NetUnknown",
				ICMPv4CodeHostUnknown:         "HostUnknown",
				ICMPv4CodeSourceIsolated:      "SourceIsolated",
				ICMPv4CodeNetAdminProhibited:  "NetAdminProhibited",
				ICMPv4CodeHostAdminProhibited: "HostAdminProhibited",
				ICMPv4CodeNetTOS:              "NetTOS",
				ICMPv4CodeHostTOS:             "HostTOS",
				ICMPv4CodeCommAdminProhibited: "CommAdminProhibited",
				ICMPv4CodeHostPrecedence:      "HostPrecedence",
				ICMPv4CodePrecedenceCutoff:    "PrecedenceCutoff",
			},
		},
		ICMPv4TypeTimeExceeded: {
			"TimeExceeded", &map[uint8]string{
				ICMPv4CodeTTLExceeded:                    "TTLExceeded",
				ICMPv4CodeFragmentReassemblyTimeExceeded: "FragmentReassemblyTimeExceeded",
			},
		},
		ICMPv4TypeParameterProblem: {
			"ParameterProblem", &map[uint8]string{
				ICMPv4CodePointerIndicatesError: "PointerIndicatesError",
				ICMPv4CodeMissingOption:         "MissingOption",
				ICMPv4CodeBadLength:             "BadLength",
			},
		},
		ICMPv4TypeSourceQuench: {
			"SourceQuench", nil,
		},
		ICMPv4TypeRedirect: {
			"Redirect", &map[uint8]string{
				ICMPv4CodeNet:     "Net",
				ICMPv4CodeHost:    "Host",
				ICMPv4CodeTOSNet:  "TOS+Net",
				ICMPv4CodeTOSHost: "TOS+Host",
			},
		},
		ICMPv4TypeEchoRequest: {
			"EchoRequest", nil,
		},
		ICMPv4TypeEchoReply: {
			"EchoReply", nil,
		},
		ICMPv4TypeTimestampRequest: {
			"TimestampRequest", nil,
		},
		ICMPv4TypeTimestampReply: {
			"TimestampReply", nil,
		},
		ICMPv4TypeInfoRequest: {
			"InfoRequest", nil,
		},
		ICMPv4TypeInfoReply: {
			"InfoReply", nil,
		},
		ICMPv4TypeRouterSolicitation: {
			"RouterSolicitation", nil,
		},
		ICMPv4TypeRouterAdvertisement: {
			"RouterAdvertisement", nil,
		},
		ICMPv4TypeAddressMaskRequest: {
			"AddressMaskRequest", nil,
		},
		ICMPv4TypeAddressMaskReply: {
			"AddressMaskReply", nil,
		},
	}
)

type ICMPv4TypeCode uint16

// Type returns the ICMPv4 type field.
func (a ICMPv4TypeCode) Type() uint8 {
	return uint8(a >> 8)
}

// Code returns the ICMPv4 code field.
func (a ICMPv4TypeCode) Code() uint8 {
	return uint8(a)
}

func (a ICMPv4TypeCode) String() string {
	t, c := a.Type(), a.Code()
	strInfo, ok := icmpv4TypeCodeInfo[t]
	if !ok {
		// Unknown ICMPv4 type field
		return fmt.Sprintf("%d(%d)", t, c)
	}
	typeStr := strInfo.typeStr
	if strInfo.codeStr == nil && c == 0 {
		// The ICMPv4 type does not make use of the code field
		return fmt.Sprintf("%s", strInfo.typeStr)
	}
	if strInfo.codeStr == nil && c != 0 {
		// The ICMPv4 type does not make use of the code field, but it is present anyway
		return fmt.Sprintf("%s(Code: %d)", typeStr, c)
	}
	codeStr, ok := (*strInfo.codeStr)[c]
	if !ok {
		// We don't know this ICMPv4 code; print the numerical value
		return fmt.Sprintf("%s(Code: %d)", typeStr, c)
	}
	return fmt.Sprintf("%s(%s)", typeStr, codeStr)
}

func (a ICMPv4TypeCode) GoString() string {
	t := reflect.TypeOf(a)
	return fmt.Sprintf("%s(%d, %d)", t.String(), a.Type(), a.Code())
}

// SerializeTo writes the ICMPv4TypeCode value to the 'bytes' buffer.
func (a ICMPv4TypeCode) SerializeTo(bytes []byte) {
	binary.BigEndian.PutUint16(bytes, uint16(a))
}

// CreateICMPv4TypeCode is a convenience function to create an ICMPv4TypeCode
// gopacket type from the ICMPv4 type and code values.
func CreateICMPv4TypeCode(typ uint8, code uint8) ICMPv4TypeCode {
	return ICMPv4TypeCode(binary.BigEndian.Uint16([]byte{typ, code}))
}

type ICMPv4 struct {
	TypeCode ICMPv4TypeCode

	Checksum uint16

	ID  uint16
	Seq uint16
}

func Parse(data []byte) (*ICMPv4, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("Incorrect ICMP header size: %d", len(data))
	}

	i := ICMPv4{}
	i.TypeCode = CreateICMPv4TypeCode(data[0], data[1])
	i.Checksum = binary.BigEndian.Uint16(data[2:4])
	i.ID = binary.BigEndian.Uint16(data[4:6])
	i.Seq = binary.BigEndian.Uint16(data[6:8])
	return &i, nil
}

func (i ICMPv4) String() string {
	return fmt.Sprintf("type=%s, checksum=%d, id=%d, seq=%d", i.TypeCode.String(), i.Checksum, i.ID, i.Seq)
}
