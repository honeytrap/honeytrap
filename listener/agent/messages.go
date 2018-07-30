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
package agent

import (
	"bytes"
	"encoding/binary"
	"net"
)

const (
	TypeHello             int = 0x00
	TypeReadWriteTCP      int = 0x01
	TypeHandshake         int = 0x02
	TypeHandshakeResponse int = 0x03
	TypeEOF               int = 0x04
	TypePing              int = 0x05
	TypeReadWriteUDP      int = 0x06
)

type Handshake struct {
	ProtocolVersion int

	CommitID      string
	ShortCommitID string

	Version string

	Token string
}

func (hs *Handshake) UnmarshalBinary(data []byte) error {
	d := NewDecoder(data)
	hs.ProtocolVersion = d.ReadUint16()
	hs.Version = d.ReadString()
	hs.ShortCommitID = d.ReadString()
	hs.CommitID = d.ReadString()
	hs.Token = d.ReadString()
	return nil
}

func (hs Handshake) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	e := NewEncoder(&buff, binary.LittleEndian)

	e.WriteUint16(hs.ProtocolVersion)
	e.WriteString(hs.Version)
	e.WriteString(hs.ShortCommitID)
	e.WriteString(hs.CommitID)

	e.WriteString(hs.Token)

	return buff.Bytes(), nil
}

type HandshakeResponse struct {
	Addresses []net.Addr
}

func (h *HandshakeResponse) UnmarshalBinary(data []byte) error {
	d := NewDecoder(data)
	n := d.ReadUint8()

	h.Addresses = make([]net.Addr, n)

	for i := 0; i < n; i++ {
		h.Addresses[i] = d.ReadAddr()
	}

	return nil
}

func (h HandshakeResponse) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	e := NewEncoder(&buff, binary.LittleEndian)

	e.WriteUint8(len(h.Addresses))

	for _, address := range h.Addresses {
		e.WriteAddr(address)
	}

	e.Flush()

	return buff.Bytes(), nil
}

type Hello struct {
	Laddr net.Addr
	Raddr net.Addr
}

func (h Hello) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	e := NewEncoder(&buff, binary.LittleEndian)

	e.WriteAddr(h.Laddr)
	e.WriteAddr(h.Raddr)

	e.Flush()

	return buff.Bytes(), nil
}

func (h *Hello) UnmarshalBinary(data []byte) error {
	decoder := NewDecoder(data)

	h.Laddr = decoder.ReadAddr()
	h.Raddr = decoder.ReadAddr()
	return nil
}

type Ping struct {
}

func (h *Ping) UnmarshalBinary(data []byte) error {
	return nil
}

func (h Ping) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}
	return buff.Bytes(), nil
}

type EOF struct {
	Laddr net.Addr
	Raddr net.Addr
}

func (e *EOF) UnmarshalBinary(data []byte) error {
	decoder := NewDecoder(data)

	e.Laddr = decoder.ReadAddr()
	e.Raddr = decoder.ReadAddr()

	return nil
}

func (e EOF) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	enc := NewEncoder(&buff, binary.LittleEndian)

	enc.WriteAddr(e.Laddr)
	enc.WriteAddr(e.Raddr)

	enc.Flush()

	return buff.Bytes(), nil
}

type ReadWriteTCP struct {
	Laddr net.Addr
	Raddr net.Addr

	Payload []byte
}

func (rw ReadWriteTCP) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	e := NewEncoder(&buff, binary.LittleEndian)

	e.WriteAddr(rw.Laddr)
	e.WriteAddr(rw.Raddr)

	e.WriteData(rw.Payload)

	e.Flush()

	return buff.Bytes(), nil
}

func (rw *ReadWriteTCP) UnmarshalBinary(data []byte) error {
	decoder := NewDecoder(data)

	rw.Laddr = decoder.ReadAddr()
	rw.Raddr = decoder.ReadAddr()

	rw.Payload = decoder.ReadData()

	return nil
}

type ReadWriteUDP struct {
	Laddr net.Addr
	Raddr net.Addr

	Payload []byte
}

func (rwu ReadWriteUDP) MarshalBinary() ([]byte, error) {
	buff := bytes.Buffer{}

	e := NewEncoder(&buff, binary.LittleEndian)

	e.WriteAddr(rwu.Laddr)
	e.WriteAddr(rwu.Raddr)

	e.WriteData(rwu.Payload)

	e.Flush()
	return buff.Bytes(), nil
}

func (rwu *ReadWriteUDP) UnmarshalBinary(data []byte) error {
	decoder := NewDecoder(data)

	rwu.Laddr = decoder.ReadAddr()
	rwu.Raddr = decoder.ReadAddr()

	rwu.Payload = decoder.ReadData()

	return nil
}
