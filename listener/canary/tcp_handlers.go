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
	"encoding/binary"
	"fmt"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/event"
	"bytes"
)

var (
	// EventCategoryTCP contains events for ssdp traffic
	EventCategoryTCP = event.Category("tcp")
)

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryHTTP = event.Category("http")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeHTTP(conn net.Conn) error {
	defer conn.Close()

	request, err := http.ReadRequest(
		bufio.NewReader(conn),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryHTTP,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.method", request.Method),
		event.Custom("http.uri", request.URL.String()),
		event.Custom("http.proto", request.Proto),
		event.Custom("http.headers", request.Header),
		event.Custom("http.host", request.Header.Get("Host")),
		event.Custom("http.content-type", request.Header.Get("Content-Type")),
		event.Custom("http.user-agent", request.Header.Get("User-Agent")),
	))

	resp := http.Response{}

	w := bufio.NewWriter(conn)
	resp.Write(w)

	w.Flush()
	_ = w

	fmt.Printf("%+v", request)
	return nil
}

var (
	// EventCategoryElasticsearch contains events for elasticsearch traffic
	EventCategoryElasticsearch = event.Category("elasticsearch")
)

// DecodeElasticsearch will decode NTP packets
func (c *Canary) DecodeElasticsearch(conn net.Conn) error {
	defer conn.Close()

	request, err := http.ReadRequest(
		bufio.NewReader(conn),
	)
	if err != nil {
		// log error / send error channel
		return nil
	}

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryElasticsearch,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Custom("http.method", request.Method),
		event.Custom("http.uri", request.URL.String()),
		event.Custom("http.proto", request.Proto),
		event.Custom("http.headers", request.Header),
		event.Custom("http.host", request.Header.Get("Host")),
		event.Custom("http.content-type", request.Header.Get("Content-Type")),
		event.Custom("http.user-agent", request.Header.Get("User-Agent")),
	))

	resp := http.Response{}

	w := bufio.NewWriter(conn)
	resp.Write(w)

	w.Flush()
	_ = w

	fmt.Printf("%+v", request)
	return nil
}

var (
	// EventCategoryHTTPS contains events for https traffic
	EventCategoryHTTPS = event.Category("https")
)

// DecodeHTTPS will decode NTP packets
func (c *Canary) DecodeHTTPS(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	offset := 0

	contentType := buff[offset]
	offset++

	version := binary.BigEndian.Uint16(buff[offset : offset+2])
	offset += 2

	options := []event.Option{
		CanaryOptions,
		EventCategoryHTTPS,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	}

	options = append(options, event.Custom("https.content-type", contentType))
	if contentType == 0x16 {
		recordLength := binary.BigEndian.Uint16(buff[offset : offset+2])
		offset += 2

		messageType := buff[offset]
		offset++

		messageLength := uint32(buff[offset])<<24 + uint32(buff[offset+1])<<16 + uint32(buff[offset+2])
		offset += 3

		clientVersion := binary.BigEndian.Uint16(buff[offset : offset+4])

		offset += 4

		random := buff[offset : offset+36]

		options = append(options, []event.Option{
			event.Custom("https.content-type", fmt.Sprintf("%x", contentType)),
			event.Custom("https.version", fmt.Sprintf("%d", version)),
			event.Custom("https.record-length", fmt.Sprintf("%d", recordLength)),
			event.Custom("https.message-type", fmt.Sprintf("%x", messageType)),
			event.Custom("https.message-length", fmt.Sprintf("%d", messageLength)),
			event.Custom("https.client-version", fmt.Sprintf("0x%x", clientVersion)),
			event.Custom("https.random", fmt.Sprintf("%x", random)),
		}...)

		if clientVersion != 0x304 {
			randomEpoch := binary.BigEndian.Uint32(buff[2:6])
			options = append(options, event.Custom("https.random-epoch", fmt.Sprintf("%d", randomEpoch)))
		}

		if v, ok := map[uint16]string{
			0x8001: "PCT_VERSION",
			0x0002: "SSLV2_VERSION",
			0x300:  "SSLV3_VERSION",
			0x301:  "TLSV1_VERSION",
			0x302:  "TLSV1DOT1_VERSION",
			0x303:  "TLSV1DOT2_VERSION",
			0x304:  "TLSV1DOT3_VERSION",
		}[clientVersion]; ok {
			options = append(options, event.Custom("https.client-version-text", v))
		}

		// add specific detections, reflection attack detection etc
	}

	c.events.Send(event.New(
		options...,
	))

	return nil
}

var (
	// EventCategoryMSSQL contains events for ssdp traffic
	EventCategoryMSSQL = event.Category("mssql")
)

// DecodeMSSQL will decode NTP packets
func (c *Canary) DecodeMSSQL(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryMSSQL,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryTelnet contains events for ssdp traffic
	EventCategoryTelnet = event.Category("telnet")
)

// DecodeTelnet will decode NTP packets
func (c *Canary) DecodeTelnet(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryTelnet,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryRedis contains events for ssdp traffic
	EventCategoryRedis = event.Category("redis")
)

// DecodeRedis will decode NTP packets
func (c *Canary) DecodeRedis(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryRedis,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryRDP contains events for ssdp traffic
	EventCategoryRDP = event.Category("rdp")
)

// DecodeRDP will decode NTP packets
func (c *Canary) DecodeRDP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryRDP,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryFTP contains events for ssdp traffic
	EventCategoryFTP = event.Category("ftp")
)

// DecodeFTP will decode NTP packets
func (c *Canary) DecodeFTP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryNBTIP,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryNBTIP contains events for ssdp traffic
	EventCategoryNBTIP = event.Category("nbt-ip")
)

// DecodeNBTIP will decode NTP packets
func (c *Canary) DecodeNBTIP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryNBTIP,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategorySMBIP contains events for ssdp traffic
	EventCategorySMBIP = event.Category("smb-ip")
)

// DecodeSMBIP will decode NTP packets
func (c *Canary) DecodeSMBIP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)
	r := bytes.NewBuffer(buff)

	options := []event.Option{
		CanaryOptions,
		EventCategorySMBIP,
		event.Protocol("tcp"),
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	}

	magicBytes := make([]byte, 4)
	r.Read(magicBytes)
	smb2Header := []byte{0xFE, byte('S'), byte('M'), byte('B')}
	if bytes.Equal(magicBytes, smb2Header) {
		// https://wiki.wireshark.org/SMB2
		options = append(options, event.Custom("smb.version", "2"))

		lengthBuf := make([]byte, 2)
		r.Read(lengthBuf)
		// length := binary.BigEndian.Uint16(lengthBuf)
		r.Next(2) // padding

		statusBuf := make([]byte, 4)
		r.Read(statusBuf)
		status := binary.BigEndian.Uint16(statusBuf)
		options = append(options, event.Custom("smb.status", fmt.Sprintf("%d", status)))

		opcodeBuf := make([]byte, 2)
		r.Read(opcodeBuf)
		opcode := binary.BigEndian.Uint16(opcodeBuf)

		if v, ok := map[uint16]string{
			0x00: "SMB2/NegotiateProtocol",
			0x01: "SMB2/SessionSetup",
			0x02: "SMB2/SessionLogoff",
			0x03: "SMB2/TreeConnect",
			0x04: "SMB2/TreeDisconnect",
			0x05: "SMB2/Create",
			0x06: "SMB2/Close",
			0x07: "SMB2/Flush",
			0x08: "SMB2/Read ",
			0x09: "SMB2/Write",
			0x0a: "SMB2/Lock ",
			0x0b: "SMB2/Ioctl",
			0x0c: "SMB2/Cancel",
			0x0d: "SMB2/KeepAlive",
			0x0e: "SMB2/Find",
			0x0f: "SMB2/Notify",
			0x10: "SMB2/GetInfo",
			0x11: "SMB2/SetInfo",
			0x12: "SMB2/Break",
		}[opcode]; ok {
			options = append(options, event.Custom("smb.opcode", fmt.Sprintf("%s", v)))
		}
	}
	// add specific detections, reflection attack detection etc

	c.events.Send(event.New(
		options...,
	))

	return nil
}
