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
package rdp

import (
	"bufio"
	"context"
	"fmt"
	"net"

	logging "github.com/op/go-logging"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
	"github.com/honeytrap/honeytrap/services"
)

const (
	TDPUConnectionRequest uint8 = 0xe0
	TPDUConnectionConfirm       = 0xd0
	TDPUData                    = 0xf0
	TPDUReject                  = 0x50
	TPDUDataAck                 = 0x60

	TypeRDPNegReq  uint8 = 1
	ProtocolRDP          = 0
	ProtocolSSL          = 1
	ProtocolHybrid       = 2

	TypeRDPNegResponse          uint8 = 2
	ExtendedClientDataSupported       = 1
	DynVCGFXProtocolSupported         = 2

	TypeRDPNegFailure               uint8 = 3
	SSLRequiredByServer                   = 1
	SSLNotAllowedByServer                 = 2
	SSLCertNotOnServer                    = 3
	InconsistentFlags                     = 4
	HybridRequiredByServer                = 5
	SSLWithUserAuthRequiredByServer       = 6
)

// https://github.com/citronneur/rdpy
// https://wiki.wireshark.org/RDP (H4)

var log = logging.MustGetLogger("services/rdp")

var (
	_ = services.Register("rdp", RDP)
)

func RDP(options ...services.ServicerFunc) services.Servicer {
	s := &rdpService{}
	for _, o := range options {
		o(s)
	}

	return s
}

type rdpService struct {
	c pushers.Channel
}

func (s *rdpService) SetChannel(c pushers.Channel) {
	s.c = c
}

type TPKT struct {
	Version  uint8
	Reserved uint8
	Length   uint16

	Payload []byte
}

func (v *TPKT) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return fmt.Errorf("Expected 4 bytes, got %d", len(data))
	}

	v.Version = uint8(data[0])
	v.Reserved = uint8(data[1])
	v.Length = uint16(data[2])>>8 + uint16(data[3])

	return nil
}

type COTP struct {
	Length  uint8
	PDUType uint8

	DestinationReference uint16
	SourceReference      uint16

	// Class
	// Extended formats
	// NoExplicitFlowControl
}

func (v *COTP) UnmarshalBinary(data []byte) error {
	if len(data) != 7 {
		return fmt.Errorf("Expected 7 bytes, got %d", len(data))
	}

	v.Length = uint8(data[0])
	v.PDUType = uint8(data[1])
	v.DestinationReference = uint16(data[2])>>8 + uint16(data[3])
	v.SourceReference = uint16(data[4])>>8 + uint16(data[5])

	return nil
}

func (s *rdpService) Handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	rdr := bufio.NewReader(conn)

	for {
		// TPKT
		b := make([]byte, 4)
		if _, err := rdr.Read(b); err != nil {
			return err
		}

		tpkt := TPKT{}
		if err := tpkt.UnmarshalBinary(b); err != nil {
			return err
		}

		// COTP
		b, err := rdr.Peek(1)
		if err != nil {
			return err
		}

		len := b[0]

		b = make([]byte, len)
		if _, err := rdr.Read(b); err != nil {
			return err
		}

		cotp := COTP{}
		if err := cotp.UnmarshalBinary(b); err != nil {
			return err
		}

		if cotp.PDUType == 0xe0 {
			b = make([]byte, 2048)
			n, err := rdr.Read(b)
			if err != nil {
				return err
			}

			b = b[:n]

			cookie := string(b)

			s.c.Send(event.New(
				event.Sensor("rdp"),
				event.Service("rdp"),
				event.Category("connect-request"),
				event.Type("connect"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("rdp.cookie", cookie),
			))

			conn.Write([]byte{0x03, 0x00, 0x00, 0x0b, 0x06, 0xd0, 0x00, 0x00, 0x12, 0x34, 0x00})
		} else if cotp.PDUType == 0xf0 /* DT Data */ {
			b = make([]byte, 2048)
			n, err := conn.Read(b)
			if err != nil {
				return err
			}

			b = b[:n]

			fmt.Printf("%x\n", b)
		} else {
			return nil
		}
	}

	return nil
}
