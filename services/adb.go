/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (adbs://dutchsec.com/)
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
* <adb://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See adbs://honeytrap.io/ for more details. All requests should be sent to
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
package services

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("adb", Adb)
)

func Adb(options ...ServicerFunc) Servicer {
	s := &adbService{}

	for _, o := range options {
		o(s)
	}

	return s
}

type adbService struct {
	c pushers.Channel
}

func (s *adbService) SetChannel(c pushers.Channel) {
	s.c = c
}

func makeAdbPacket(command, arg1, arg2, info []byte) []byte {
	hdr := make([]byte, 24)
	copy(hdr[0:4], command)
	copy(hdr[4:8], arg1)
	copy(hdr[8:12], arg2)
	binary.LittleEndian.PutUint32(hdr[12:16], uint32(len(info)))
	// See transport.c in adb
	var crc uint32
	for _, b := range info {
		crc += uint32(b)
	}
	binary.LittleEndian.PutUint32(hdr[16:20], crc)
	hdr[20] = command[0] ^ 0xff
	hdr[21] = command[1] ^ 0xff
	hdr[22] = command[2] ^ 0xff
	hdr[23] = command[3] ^ 0xff
	return append(hdr, info...)
}

func (s *adbService) Handle(ctx context.Context, conn net.Conn) error {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		panic(err)
	}
	payload := buf[:n]
	cmd := payload[:4]
	if !bytes.Equal(cmd, []byte("CNXN")) {
		log.Errorf("Expected CNXN, got %s", string(cmd))
		return nil
	}
	s.c.Send(event.New(
		EventOptions,
		event.Category("adb"),
		event.Type("connection"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(payload[24:]),
	))
	/* Android >=4.4 has an additional authentication step, where the hosts exchange keys.
	 * This is skipped here (pretending to be Android <4.4) for ease of implementation, both client- and server-side.
	 */
	conn.Write(makeAdbPacket(
		[]byte("CNXN"),
		// version
		[]byte{0x00, 0x00, 0x00, 0x01},
		// max length
		[]byte{0x00, 0x10, 0x00, 0x00},
		// Galaxy S8 phone: https://github.com/pytorch/cpuinfo/blob/master/test/build.prop/galaxy-s8-global.log
		[]byte("device::ro.product.name=dreamltexx;ro.product.model=SM-G950F;ro.product.device=dreamlte;\x00"),
	))
	var commandBuffer []byte
	localId := []byte{0x09, 0x00, 0x00, 0x00}
	var remoteId []byte
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			panic(err)
		}
		payload = buf[:n]
		cmd = payload[:4]
		if bytes.Equal(cmd, []byte("OPEN")) {
			// Open a shell
			s.c.Send(event.New(
				EventOptions,
				event.Category("adb"),
				event.Type("connection"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Payload(payload[24:]),
			))
			remoteId = payload[4:8]
			conn.Write(makeAdbPacket([]byte("OKAY"), localId, remoteId, []byte{}))
			conn.Write(makeAdbPacket([]byte("WRTE"), localId, remoteId, []byte("shell@SWDG4522:/ $ ")))
		} else if bytes.Equal(cmd, []byte("WRTE")) {
			// Receive data
			response := payload[24:]
			commandBuffer = append(commandBuffer, response...)
			conn.Write(makeAdbPacket([]byte("OKAY"), localId, remoteId, []byte{}))
			if bytes.ContainsRune(commandBuffer, '\r') {
				s.c.Send(event.New(
					EventOptions,
					event.Category("adb"),
					event.Type("command"),
					event.SourceAddr(conn.RemoteAddr()),
					event.DestinationAddr(conn.LocalAddr()),
					event.Payload(commandBuffer),
				))
				response = append(response, []byte("\r\nshell@SWDG4522:/ $ ")...)
				commandBuffer = []byte{}
			}
			conn.Write(makeAdbPacket([]byte("WRTE"), localId, remoteId, response))
		} else if bytes.Equal(cmd, []byte("OKAY")) {
		} else if bytes.Equal(cmd, []byte("CLSE")) {
			return nil
		} else {
			log.Warningf("Received unknown command %s", string(cmd))
			conn.Write(makeAdbPacket([]byte("OKAY"), localId, remoteId, []byte{}))
		}
	}
	return nil
}
