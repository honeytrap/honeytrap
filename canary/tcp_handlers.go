package canary

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"

	"github.com/honeytrap/honeytrap/pushers/event"
)

var (
	// EventCategoryTCP contains events for ssdp traffic
	EventCategoryTCP = event.Category("tcp")
)

// EventTCPPayload will return a snmp event struct
func EventTCPPayload(src, dst net.IP, srcport, dstport uint16, payload []byte) event.Event {
	return event.New(
		CanaryOptions,
		EventCategoryTCP,
		event.ServiceStarted,
		event.SourceIP(src),
		event.DestinationIP(dst),
		event.SourcePort(srcport),
		event.DestinationPort(dstport),
		event.Payload(payload),
	)
}

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
		event.ServiceStarted,
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
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryHTTPS = event.Category("https")
)

// DecodeHTTP will decode NTP packets
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
		event.ServiceStarted,
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

		if clientVersion == 0x304 {
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
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryMSSQL = event.Category("mssql")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeMSSQL(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryMSSQL,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryTelnet = event.Category("telnet")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeTelnet(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryTelnet,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryRedis = event.Category("redis")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeRedis(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryRedis,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryRDP = event.Category("rdp")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeRDP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryRDP,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryFTP = event.Category("ftp")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeFTP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryNBTIP,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategoryNBTIP = event.Category("nbt-ip")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeNBTIP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	// add specific detections, reflection attack detection etc
	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryNBTIP,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	))

	return nil
}

var (
	// EventCategoryHTTP contains events for ssdp traffic
	EventCategorySMBIP = event.Category("smb-ip")
)

// DecodeHTTP will decode NTP packets
func (c *Canary) DecodeSMBIP(conn net.Conn) error {
	defer conn.Close()

	buff := make([]byte, 2048)
	n, _ := conn.Read(buff)

	options := []event.Option{
		CanaryOptions,
		EventCategorySMBIP,
		event.ServiceStarted,
		event.SourceAddr(conn.RemoteAddr()),
		event.DestinationAddr(conn.LocalAddr()),
		event.Payload(buff[:n]),
	}

	offset := 0
	if buff[0] == 0xFE {
		// https://wiki.wireshark.org/SMB2
		options = append(options, event.Custom("smb.version", "2"))

		offset++

		length := binary.BigEndian.Uint16(buff[offset : offset+2])
		offset += 4
		_ = length

		status := binary.BigEndian.Uint16(buff[offset : offset+4])
		offset += 4
		options = append(options, event.Custom("smb.status", fmt.Sprintf("%d", status)))

		opcode := binary.BigEndian.Uint16(buff[offset : offset+2])
		offset += 4

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
