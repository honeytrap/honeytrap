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
package services

import (
	"bufio"
	"context"
	"encoding/hex"
	"net"

	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/pushers"
)

var (
	_ = Register("tftp", TFTP)
)

func TFTP(options ...ServicerFunc) Servicer {
	s := &tftpService{
		limiter: NewLimiter(),
	}
	for _, o := range options {
		o(s)
	}
	s.buffers = make(map[string]*tftpFile)
	return s
}

type tftpFile struct {
	filename string
	mode     string
	content  []byte
}

type tftpService struct {
	ch pushers.Channel

	limiter *Limiter

	buffers map[string]*tftpFile
}

func (s *tftpService) SetChannel(c pushers.Channel) {
	s.ch = c
}

func (s *tftpService) Handle(ctx context.Context, conn net.Conn) error {
	if conn.RemoteAddr().Network() == "udp" {
		/* Selectively drop packets to prevent amplification attacks. This is a
		 * simple approach that "just works", since for each client packet there
		 * is one response packet from the server
		 */
		if !s.limiter.Allow(conn.RemoteAddr()) {
			return nil
		}
	} else {
		log.Error("Expected UDP connection, got %s", conn.RemoteAddr().Network())
	}

	var (
		RRQ   = 1
		WRQ   = 2
		DATA  = 3
		ACK   = 4
		ERROR = 5
	)

	b := bufio.NewReader(conn)
	packetType := make([]byte, 2)
	if _, err := b.Read(packetType); err != nil {
		return err
	}
	switch int(packetType[1]) { // The first byte is always zero at the time of writing
	case RRQ:
		filename, err := b.ReadString(byte(0))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		mode, err := b.ReadString(byte(0))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		s.ch.Send(event.New(
			EventOptions,
			event.Category("tftp"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("tftp-read"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("tftp.filename", filename),
			event.Custom("tftp.mode", mode),
		))
		// Discard extensions
		/*
			for {
				ext, err := b.ReadString(byte(0))
				if err != nil {
					log.Error(err.Error())
					break
				}
			}
		*/
		// Return an ERROR for "file not found", to avoid amplification attacks
		message := []byte{
			0x00, byte(ERROR), // Type
			0x00, 0x01, // Error code
			0x00, // Error description (empty)
		}
		conn.Write(message)
	case WRQ:
		filename, err := b.ReadString(byte(0))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		mode, err := b.ReadString(byte(0))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		s.ch.Send(event.New(
			EventOptions,
			event.Category("tftp"),
			event.Protocol(conn.RemoteAddr().Network()),
			event.Type("tftp-write"),
			event.SourceAddr(conn.RemoteAddr()),
			event.DestinationAddr(conn.LocalAddr()),
			event.Custom("tftp.filename", filename),
			event.Custom("tftp.mode", mode),
		))
		message := []byte{
			0x00, byte(ACK),
			0x00, 0x00,
		}
		conn.Write(message)
		addr := conn.RemoteAddr().String()
		s.buffers[addr] = &tftpFile{filename: filename, mode: mode}
	case DATA:
		blkNum := make([]byte, 2)
		if _, err := b.Read(blkNum); err != nil {
			log.Error(err.Error())
			return err
		}
		buffer := make([]byte, 512)
		n, err := b.Read(buffer)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		addr := conn.RemoteAddr().String()
		if _, ok := s.buffers[addr]; !ok {
			log.Error("DATA packet with no matching buffer!")
			message := []byte{0x00, byte(ERROR), 0x00, 0x04, 0x00}
			conn.Write(message)
			return nil
		}
		s.buffers[addr].content = append(s.buffers[addr].content, buffer[:n]...)
		message := []byte{
			0x00, byte(ACK),
			blkNum[0], blkNum[1],
		}
		conn.Write(message)
		if n != 512 { // Termination
			file := s.buffers[addr]
			delete(s.buffers, addr)
			s.ch.Send(event.New(
				EventOptions,
				event.Category("tftp"),
				event.Protocol(conn.RemoteAddr().Network()),
				event.Type("tftp-write-file"),
				event.SourceAddr(conn.RemoteAddr()),
				event.DestinationAddr(conn.LocalAddr()),
				event.Custom("tftp.filename", file.filename),
				event.Custom("tftp.mode", file.mode),
				event.Custom("tftp.file", file.content),
				event.Custom("tftp.file-hex", hex.EncodeToString(file.content)),
			))
		}
	case ACK:
		/*
			blkNum := make([]byte, 2)
			_, err := b.Read(blkNum)
			if err != nil {
				log.Error(err.Error())
				return err
			}
		*/
	case ERROR:
		// We shouldn't receive ERROR, as a server.
		log.Error("Unexpected ERROR packet (I should be a server!)")
	default:
		log.Error("Unknown packet type %X", packetType)
	}
	return nil
}
