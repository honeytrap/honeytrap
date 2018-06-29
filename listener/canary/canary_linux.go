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
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/glycerine/rbuf"
	"github.com/honeytrap/honeytrap/event"
	"github.com/honeytrap/honeytrap/listener"
	"github.com/honeytrap/honeytrap/listener/canary/arp"
	"github.com/honeytrap/honeytrap/listener/canary/ethernet"
	"github.com/honeytrap/honeytrap/listener/canary/icmp"
	"github.com/honeytrap/honeytrap/listener/canary/ipv4"
	"github.com/honeytrap/honeytrap/listener/canary/tcp"
	"github.com/honeytrap/honeytrap/listener/canary/udp"
	"github.com/honeytrap/honeytrap/pushers"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("listeners/raw")

var (
	_ = listener.Register("raw", New)
)

var (
	// EventCategoryARP
	EventCategoryARP = event.Category("arp")
)

// first dns
// ntp
// send reset?
// udp check connect or answer
// parameters
// clean up old states
// check ring buffer
// use sockets and io.Reader
// parameters: ports to include exclude/ filter (or do we want to filter the events)
// answer with data

const (
	// MaxEpollEvents defines maximum number of poll events to retrieve at once
	MaxEpollEvents = 2048
	// DefaultBufferSize defines size of receive buffer
	DefaultBufferSize = 65535
)

const (
	// EthernetTypeIPv4 is the protocol number for IPv4 traffic
	EthernetTypeIPv4 = 0x0800
	// EthernetTypeIPv6 is the protocol number for IPv6 traffic
	EthernetTypeIPv6 = 0x86DD
	// EthernetTypeARP is the protocol number for ARP traffic
	EthernetTypeARP = 0x0806
)

// Protocol specifies the network protocol
type Protocol int

const (
	// ProtocolTCP specifies tcp protocol
	ProtocolTCP Protocol = iota
	// ProtocolUDP specifies udp protocol
	ProtocolUDP
	// ProtocolICMP specifies icmp protocol
	ProtocolICMP
)

// Canary contains the canary struct
type Canary struct {
	rt RouteTable

	Interfaces []string `toml:"interfaces"`

	ch chan net.Conn

	ac ARPCache

	epfd int

	m sync.Mutex

	r *rand.Rand

	knockChan chan interface{}

	networkInterfaces []net.Interface

	events pushers.Channel

	descriptors map[string]int32

	buffer *rbuf.FixedSizeRingBuf

	stateTable StateTable

	// txqueue *Queue  <unused>
}

/*
// Queue contains packets for delivery. (currently unused)
type Queue struct {
	packets []interface{}
}
*/

// Taken from https://github.com/xiezhenye/harp/blob/master/src/arp/arp.go#L53
func htons(n uint16) uint16 {
	var (
		high = n >> 8
		ret  = n<<8 + high
	)

	return ret
}

// handleUDP will handle udp packets
func (c *Canary) handleUDP(eh *ethernet.Frame, iph *ipv4.Header, data []byte) error {
	hdr, err := udp.Unmarshal(data)
	if err != nil {
		return nil
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	// check if we have udp listeners on specified port, and answer otherwise
	// parse udp
	// we should check if the received packet is a response or request
	// detect if our interface initiated or portscan

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)

				message := event.Message("%+v", r)
				if err, ok := r.(error); ok {
					message = event.Message("%+v", err)
				}

				c.events.Send(event.New(
					CanaryOptions,
					EventCategoryTCP,
					event.SeverityFatal,

					event.SourceHardwareAddr(eh.Source),
					event.DestinationHardwareAddr(eh.Destination),

					event.SourceIP(iph.Src),
					event.DestinationIP(iph.Dst),
					event.SourcePort(hdr.Source),
					event.DestinationPort(hdr.Destination),
					event.Stack(),
					message,
					// event.Payload(buff[:n]),
				))
			}
		}()

		handlers := map[uint16]func(*ipv4.Header, *udp.Header) error{
			53:   c.DecodeDNS,
			123:  c.DecodeNTP,
			1900: c.DecodeSSDP,
			5060: c.DecodeSIP,
			161:  c.DecodeSNMP,
			162:  c.DecodeSNMPTrap,
		}

		if fn, ok := handlers[hdr.Destination]; !ok {
			// default handler
			c.knockChan <- KnockUDPPort{
				SourceHardwareAddr:      eh.Source,
				DestinationHardwareAddr: eh.Destination,
				SourceIP:                iph.Src,
				DestinationIP:           iph.Dst,
				DestinationPort:         hdr.Destination,
			}

			// do we only want to detect scans? Or also detect payloads?
			c.events.Send(event.New(
				SensorCanary,
				EventCategoryUDP,

				event.Protocol("udp"),
				event.SourceHardwareAddr(eh.Source),
				event.DestinationHardwareAddr(eh.Destination),

				event.SourceIP(iph.Src),
				event.DestinationIP(iph.Dst),

				event.SourcePort(hdr.Source),
				event.DestinationPort(hdr.Destination),

				event.Payload(hdr.Payload),
			))

		} else if err := fn(iph, hdr); err != nil {
			fmt.Printf("Could not decode udp packet: %s", err)
			// return err
			// todo to error channel
		}
	}()

	return nil
}

// handleICMP will handle tcp packets
func (c *Canary) handleICMP(eh *ethernet.Frame, iph *ipv4.Header, data []byte) error {
	_, err := icmp.Parse(data)
	if err != nil {
		return err
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	c.knockChan <- KnockICMP{
		SourceHardwareAddr:      eh.Source,
		DestinationHardwareAddr: eh.Destination,
		SourceIP:                iph.Src,
		DestinationIP:           iph.Dst,
	}

	return nil
}

// handleARP will handle arp packets
func (c *Canary) handleARP(data []byte) error {
	arp, err := arp.Parse(data)
	if err != nil {
		return err
	}

	c.events.Send(event.New(
		CanaryOptions,
		EventCategoryARP,
		event.DestinationHardwareAddr(arp.TargetMAC),
		event.SourceIP(arp.SenderIP),
		event.DestinationIP(arp.TargetIP),
		event.SourceHardwareAddr(arp.SenderMAC),
		event.DestinationHardwareAddr(arp.TargetMAC),
		event.SourceIP(arp.SenderIP),
		event.DestinationIP(arp.TargetIP),
		event.Custom("arp-sender-hardware-address", arp.SenderHardwareAddress),
		event.Custom("arp-target-hardware-address", arp.TargetHardwareAddress),
		event.Custom("arp-sender-protocol-address", arp.SenderProtocolAddress),
		event.Custom("arp-target-protocol-address", arp.TargetProtocolAddress),
		event.Custom("arp-opcode", arp.Opcode),
		event.Custom("arp-hardware-type", arp.HardwareType),
		event.Custom("arp-hardware-size", arp.HardwareSize),
		event.Custom("arp-protocol-type", arp.ProtocolType),
		event.Custom("arp-protocol-size", arp.ProtocolSize),
		event.Payload(data),
	))

	return nil
}

// isMe returns if the ip is one of our interfaces addresses
func (c *Canary) isMe(ip net.IP) bool {
	for _, intf := range c.networkInterfaces {
		addrs, _ := intf.Addrs()

		for _, addr := range addrs {
			if ip4net, ok := addr.(*net.IPNet); !ok {
			} else if ip4net.IP.Equal(ip) {
				return true
			}
		}
	}

	return false
}

// handleTCP will handle tcp packets
func (c *Canary) handleTCP(eh *ethernet.Frame, iph *ipv4.Header, data []byte) error {
	hdr, err := tcp.UnmarshalWithChecksum(data, iph.Dst, iph.Src)
	if err == tcp.ErrInvalidChecksum {
		// we are ignoring invalid checksums for now
	} else if err != nil {
		return err
	}

	if !c.isMe(iph.Dst) {
		return nil
	}

	if hdr.Source == 22 || hdr.Destination == 22 {
		return nil
	}

	state := c.stateTable.Get(iph.Src, iph.Dst, hdr.Source, hdr.Destination)
	if hdr.HasFlag(tcp.SYN) && !hdr.HasFlag(tcp.ACK) {
		// no state found
		state = c.NewState(iph.Src, hdr.Source, iph.Dst, hdr.Destination)
		state.State = SocketListen
		c.stateTable.Add(state)

		// or is state == socket?

		// new socket
		state.socket = state.NewSocket(
			&net.TCPAddr{
				IP:   iph.Src,
				Port: int(hdr.Source),
			},
			&net.TCPAddr{
				IP:   iph.Dst,
				Port: int(hdr.Destination),
			},
		)
	}

	if state == nil {
		// no existing state found, returning
		return nil // ErrNoExistingStateFound()
	}

	// this is far from ideal, but basically we want to prevent changing state values
	// in race conditions (eg from socket) and from handleTCP function
	state.m.Lock()
	defer state.m.Unlock()

	state.t = time.Now()

	// https://tools.ietf.org/html/rfc793
	// page 65

	if state.State == SocketListen {
		switch {
		case hdr.HasFlag(tcp.SYN):
			state.SendUnacknowledged = state.InitialSendSequenceNumber
			state.SendNext = state.InitialSendSequenceNumber + 1

			state.RecvNext = hdr.SeqNum
			state.RecvNext++
			c.send(state, []byte{}, tcp.SYN|tcp.ACK)
			state.SendNext++
			state.State = SocketSynReceived
			return nil
		}
	}

	// check sequence number

	switch {
	case hdr.HasFlag(tcp.RST):
		if state.State == SocketSynReceived {
			// enter listen state
			state.State = SocketListen
			return nil
		}

		switch state.State {
		case SocketEstablished:
		case SocketFinWait1:
		case SocketFinWait2:
		case SocketCloseWait:
			// If the RST bit is set then, any outstanding RECEIVEs and SEND
			// should receive "reset" responses.  All segment queues should be
			// flushed.  Users should also receive an unsolicited general
			// "connection reset" signal.  Enter the CLOSED state, delete the
			// TCB, and return.
			state.State = SocketClosed

			c.stateTable.Remove(state)
			return nil
		case SocketClosing:
		case SocketLastAck:
		case SocketTimeWait:
			// If the RST bit is set then, enter the CLOSED state, delete the
			// TCB, and return.
			state.State = SocketClosed

			c.stateTable.Remove(state)
			return nil
		}
	}

	// check security and precedence

	/*
			if state.State == SynReceived {
				// enter listen state
				return nil
			}

			state.RecvNext++
			c.ack(state, tcp.RST|tcp.ACK)
		case hdr.HasFlag(tcp.FIN):
			state.RecvNext++
			c.ack(state, tcp.FIN|tcp.ACK)
	*/
	if hdr.HasFlag(tcp.SYN) {
		// If the SYN is in the window it is an error, send a reset, any
		// outstanding RECEIVEs and SEND should receive "reset" responses,
		// all segment queues should be flushed, the user should also
		// receive an unsolicited general "connection reset" signal, enter
		// the CLOSED state, delete the TCB, and return.

		// If the SYN is not in the window this step would not be reached
		// and an ack would have been sent in the first step (sequence
		// number check).
		return nil
	}

	// check the ACK field
	if !hdr.HasFlag(tcp.ACK) {
		// if the ACK bit is off drop the segment and return
		return nil
	}

	if state.State == SocketClosing {
		state.State = SocketTimeWait
	}

	if state.State == SocketCloseWait {
		c.stateTable.Remove(state)
	}

	if state.State == SocketSynReceived {
		if state.SendUnacknowledged <= hdr.AckNum &&
			hdr.AckNum <= state.SendNext {
			state.State = SocketEstablished
		} else {
			// If the segment acknowledgment is not acceptable, form a
			// reset segment,
			// <SEQ=SEG.ACK><CTL=RST>
			// and send it.
			return nil
		}

		// listen handler,
		// linux works with syn queue
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("Recovered from panic: %+v", r)

					message := event.Message("%+v", r)
					if err, ok := r.(error); ok {
						message = event.Message("%+v", err)
					}

					c.events.Send(event.New(
						CanaryOptions,
						EventCategoryTCP,
						event.SeverityFatal,
						event.SourceHardwareAddr(eh.Source),
						event.DestinationHardwareAddr(eh.Destination),
						event.SourceIP(state.SrcIP),
						event.DestinationIP(state.DestIP),
						event.SourcePort(state.SrcPort),
						event.DestinationPort(state.DestPort),
						event.Stack(),
						message,
						// event.Payload(buff[:n]),
					))
				}
			}()

			// we can check als for signaturs to use specifiec protocol
			handlers := map[uint16]func(net.Conn) error{
				23:   c.DecodeTelnet,
				80:   c.DecodeHTTP,
				443:  c.DecodeHTTPS,
				139:  c.DecodeNBTIP,
				445:  c.DecodeSMBIP,
				1433: c.DecodeMSSQL,
				6379: c.DecodeRedis,
				9200: c.DecodeElasticsearch,
			}

			if fn, ok := handlers[hdr.Destination]; !ok {
				buff := make([]byte, 2048)

				rdr := state.socket // io.TeeReader(state.socket, os.Stdout)
				n, _ := rdr.Read(buff)

				w := bufio.NewWriter(state.socket)
				//w.WriteString("test")
				w.Flush()

				state.socket.Close()

				c.events.Send(event.New(
					CanaryOptions,
					EventCategoryTCP,
					event.ServiceStarted,
					event.Protocol("tcp"),
					event.SourceHardwareAddr(eh.Source),
					event.DestinationHardwareAddr(eh.Destination),
					event.SourceIP(state.SrcIP),
					event.DestinationIP(state.DestIP),
					event.SourcePort(state.SrcPort),
					event.DestinationPort(state.DestPort),
					event.Payload(buff[:n]),
				))

			} else if err := fn(state.socket); err != nil {
				_ = fn
			}
		}()
	}

	// SocketEstablished
	// If SND.UNA < SEG.ACK =< SND.NXT then, set SND.UNA <- SEG.ACK.

	if state.SendUnacknowledged <= hdr.AckNum &&
		hdr.AckNum <= state.SendNext {
		state.SendUnacknowledged = hdr.AckNum
	}

	// Any segments on the retransmission queue which are thereby
	// entirely acknowledged are removed.  Users should receive
	// positive acknowledgments for buffers which have been SENT and
	// fully acknowledged (i.e., SEND buffer should be returned with
	// 	"ok" response).

	// retransmission queue

	// If the ACK is a duplicate
	// (SEG.ACK < SND.UNA), it can be ignored.
	if hdr.AckNum < state.SendUnacknowledged {
	}

	// If the ACK acks
	// something not yet sent (SEG.ACK > SND.NXT) then send an ACK,
	// drop the segment, and return.
	if hdr.AckNum > state.SendUnacknowledged {
		// is this necessary?
		// c.send(state, []byte{}, tcp.ACK)
		// return nil
	}

	// If SND.UNA < SEG.ACK =< SND.NXT, the send window should be
	// updated.  If (SND.WL1 < SEG.SEQ or (SND.WL1 = SEG.SEQ and
	// 	SND.WL2 =< SEG.ACK)), set SND.WND <- SEG.WND, set
	// SND.WL1 <- SEG.SEQ, and set SND.WL2 <- SEG.ACK.

	// 	Note that SND.WND is an offset from SND.UNA, that SND.WL1
	// records the sequence number of the last segment used to update
	// SND.WND, and that SND.WL2 records the acknowledgment number of
	// the last segment used to update SND.WND.  The check here
	// prevents using old segments to update the window.

	if state.State == SocketFinWait1 {
		// In addition to the processing for the ESTABLISHED state, if
		// our FIN is now acknowledged then enter FIN-WAIT-2 and continue
		// processing in that state.
		state.State = SocketFinWait2
	} else if state.State == SocketFinWait2 {
		state.State = SocketTimeWait
	}

	if state.State == SocketEstablished ||
		state.State == SocketFinWait1 ||
		state.State == SocketFinWait2 {

		state.socket.write(hdr.Payload)

		switch {
		case hdr.HasFlag(tcp.PSH):
			state.socket.flush()
		}

		// Once the TCP takes responsibility for the data it advances
		// RCV.NXT over the data accepted, and adjusts RCV.WND as
		// apporopriate to the current buffer availability.  The total of
		// RCV.NXT and RCV.WND should not be reduced.
		state.RecvNext += uint32(len(hdr.Payload))
		// state.ReceiveWindow = state.socket.bufferavailable() //  hdr.Payload

		// TODO: not in spec, but what should we do? otherwise
		// it will ACK to the SYN, SYNACK, ACK
		if len(hdr.Payload) > 0 {
			// This acknowledgment should be piggybacked on a segment being
			// transmitted if possible without incurring undue delay.
			// fmt.Printf("ACK'ing %d %d\n", state.SendNext, state.RecvNext)
			c.send(state, []byte{}, tcp.ACK)
		}
	}

	if hdr.Ctrl&tcp.SYN == tcp.SYN {
		c.knockChan <- KnockTCPPort{
			SourceHardwareAddr:      eh.Source,
			DestinationHardwareAddr: eh.Destination,
			SourceIP:                iph.Src,
			DestinationIP:           iph.Dst,
			DestinationPort:         hdr.Destination,
		}
	}

	if hdr.Ctrl&tcp.FIN == tcp.FIN {
		// If the FIN bit is set, signal the user "connection closing" and
		// return any pending RECEIVEs with same message, advance RCV.NXT
		// over the FIN, and Option an acknowledgment for the FIN.  Note that
		// FIN implies PUSH for any segment text not yet delivered to the
		// user.
		state.RecvNext = hdr.SeqNum

		if state.State == SocketSynReceived || state.State == SocketEstablished {
			// Enter the CLOSE-WAIT state.
			state.socket.flush()
			state.socket.close()

			state.RecvNext++
			// c.send(state, []byte{}, tcp.SYN|tcp.ACK)
			c.send(state, []byte{}, tcp.FIN|tcp.ACK)
			state.SendNext++

			state.State = SocketCloseWait

			// 			state.socket.close()
		} else if state.State == SocketFinWait1 {
			// If our FIN has been ACKed (perhaps in this segment), then
			// enter TIME-WAIT, start the time-wait timer, turn off the other
			// timers; otherwise enter the CLOSING state.
			state.State = SocketClosing
		} else if state.State == SocketFinWait2 {
			state.RecvNext++

			// fmt.Printf("(socketfinwait)ACK'ing %d %d\n", state.SendNext, state.RecvNext)
			c.send(state, []byte{}, tcp.ACK)

			// Enter the TIME-WAIT state.  Start the time-wait timer, turn
			// off the other timers.
			state.State = SocketTimeWait
		}

		// we should only close when FIN but not FIN-ACK
		// set state status s.FINWAIT
		// state.socket.close()
	} else {
		// remove states
		// FIN / RST
		return nil
	}
	/*
		if hdr.Ctrl&tcp.RST == tcp.RST {
			// we should only close when RST but not RST-ACK
			state.socket.close()
		} else if hdr.Ctrl&tcp.FIN == tcp.FIN {
			// we should only close when FIN but not FIN-ACK
			// set state status s.FINWAIT
			state.socket.close()
		} else {
			// remove states
			// FIN / RST
			return nil
		}
	*/
	// check if we have tcp listeners on specified port, and answer otherwise
	return nil
}

func (c *Canary) send(state *State, payload []byte, flags tcp.Flag) error {
	// fmt.Printf("Sending packet flags=%d state=%d payload-length=%d\n%s\n", flags, state.State, len(payload), string(debug.Stack()))

	th := &tcp.Header{
		Source:      state.DestPort,
		Destination: state.SrcPort,
		SeqNum:      state.SendNext,
		AckNum:      state.RecvNext,
		Reserved:    0,
		ECN:         0,
		Ctrl:        flags,
		Window:      state.ReceiveWindow,
		Checksum:    0,
		Urgent:      0,
		Options:     []tcp.Option{},
		Payload:     payload,
	}

	data1, err := th.Marshal()
	if err != nil {
		return err
	}

	// ack the received packet
	iph := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		Flags:    0,
		FragOff:  0,
		TTL:      128,
		Src:      state.DestIP,
		Dst:      state.SrcIP,
		ID:       int(state.ID), // state.ID() which will increment automatically
		Protocol: 6,
		TotalLen: 20 + len(data1),
	}

	data, err := iph.Marshal()
	if err != nil {
		return err
	}

	state.ID++

	updateTCPChecksum(iph, data1)

	data = append(data, data1...)

	csum := uint32(0)

	// calculate correct ip header length here.
	length := 20

	// calculate options?
	for i := 0; i < length; i += 2 {
		if i == 10 {
			continue
		}

		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}

	for {
		// Break when sum is less or equals to 0xFFFF
		if csum <= 65535 {
			break
		}
		// Add carry to the sum
		csum = (csum >> 16) + uint32(uint16(csum))
	}

	csum = uint32(^uint16(csum))

	data[10] = uint8((csum >> 8) & 0xFF)
	data[11] = uint8(csum & 0xFF)

	// Src := net.IPv4(data1[12], data1[13], data1[14], data1[15])
	dst := net.IPv4(data[16], data[17], data[18], data[19])

	ae := c.ac.Get(dst)
	if ae == nil {
		// TODO(make function)
		for _, route := range c.rt {

			// find shortest route
			if !route.Destination.Contains(dst) {
				continue
			}

			ae = c.ac.Get(route.Gateway)
			break
		}

	}

	ef := ethernet.Frame{
		Source:      c.networkInterfaces[0].HardwareAddr,
		Destination: ae.HardwareAddress,
		Type:        0x0800,
	}

	data2, err := ef.Marshal()
	if err != nil {
		fmt.Println("Error marshalling ethernet frame: ", err)
		return err
	}

	data = append(data2, data...)

	c.buffer.Write([]byte{byte((len(data) & 0xFF00) >> 8), byte(len(data) & 0xFF)})
	c.buffer.Write(data)

	fd := c.descriptors[ae.Interface]

	// copy to retransmission queue
	/*
		c.txqueue = append(c.txqueue, Packet{
			t: time.Now()
			d:
		})
	*/

	return syscall.EpollCtl(c.epfd, syscall.EPOLL_CTL_MOD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLOUT,
		Fd:     int32(fd),
	})
}

// Count occurrences in s of any bytes in t.
func countAnyByte(s string, t string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			n++
		}
	}
	return n
}

// Split s at any bytes in t.
func splitAtBytes(s string, t string) []string {
	a := make([]string, 1+countAnyByte(s, t))
	n := 0
	last := 0
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(t, s[i]) >= 0 {
			if last < i {
				a[n] = s[last:i]
				n++
			}
			last = i + 1
		}
	}
	if last < len(s) {
		a[n] = s[last:]
		n++
	}
	return a[0:n]
}

// New will return a Canary for specified interfaces. Events will be delivered through
// events
func New(options ...func(listener.Listener) error) (listener.Listener, error) {
	ch := make(chan net.Conn)

	rt, err := parseRouteTable("/proc/net/route")
	if err != nil {
		return nil, fmt.Errorf("Could not parse route table: %s", err.Error())
	}

	ac, err := parseARPCache("/proc/net/arp")
	if err != nil {
		return nil, fmt.Errorf("Could not parse arp cache: %s", err.Error())
	}

	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("epoll_create1: %s", err.Error())
	}

	// todo: cleanup statetable

	networkInterfaces := []net.Interface{}
	descriptors := map[string]int32{}

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	l := &Canary{
		ac:                ac,
		rt:                rt,
		epfd:              epfd,
		descriptors:       descriptors,
		networkInterfaces: networkInterfaces,
		r:                 r,
		knockChan:         make(chan interface{}, 100),
		events:            pushers.MustDummy(),
		m:                 sync.Mutex{},
		ch:                ch,
		buffer:            rbuf.NewFixedSizeRingBuf(65535),
	}

	for _, option := range options {
		option(l)
	}

	for _, name := range l.Interfaces {
		intf, err := net.InterfaceByName(name)
		if err != nil {
			return nil, err
		}

		fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		if err != nil {
			return nil, fmt.Errorf("Could not create socket: %s", err.Error())
		}

		if fd < 0 {
			return nil, fmt.Errorf("Socket error: return < 0")
		}

		if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLERR, /*| syscall.EPOLL_NONBLOCK*/
			Fd:     int32(fd),
		}); err != nil {
			return nil, fmt.Errorf("epollctl: %s", err.Error())
		}

		l.descriptors[intf.Name] = int32(fd)
		l.networkInterfaces = append(l.networkInterfaces, *intf)
	}

	return l, nil
}

func (c *Canary) SetChannel(ch pushers.Channel) {
	c.events = ch
}

func (c *Canary) Accept() (net.Conn, error) {
	conn := <-c.ch
	return conn, nil
}

// Close will close the canary
func (c *Canary) Close() {
	syscall.Close(c.epfd)
}

func updateTCPChecksum(iph *ipv4.Header, data []byte) {
	length := len(data)

	csum := uint32(0)

	csum += (uint32(iph.Src[12]) + uint32(iph.Src[14])) << 8
	csum += uint32(iph.Src[13]) + uint32(iph.Src[15])
	csum += (uint32(iph.Dst[12]) + uint32(iph.Dst[14])) << 8
	csum += uint32(iph.Dst[13]) + uint32(iph.Dst[15])

	csum += uint32(6)
	csum += uint32(length) & 0xffff
	csum += uint32(length) >> 16

	length = len(data) - 1

	// calculate correct ip header length here.
	for i := 0; i < length; i += 2 {
		if i == 16 {
			continue
		}

		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}

	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}

	csum = uint32(^uint16(csum + (csum >> 16)))

	data[16] = uint8((csum >> 8) & 0xFF)
	data[17] = uint8(csum & 0xFF)
}

// send will queue a packet for sending
func (c *Canary) transmit(fd int32) error {
	for {
		buff := [2]byte{}

		_, err := c.buffer.ReadAndMaybeAdvance(buff[:], true)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading buffer 1: %s", err)
			return err
		}

		len := uint32(buff[0])<<8 + uint32(buff[1])

		buffer := make([]byte, len)
		n, err := c.buffer.Read(buffer)
		if err != nil {
			log.Errorf("Error reading buffer 2: %s", err)
			return err
		}

		to := &syscall.SockaddrLinklayer{
			Protocol: htons(syscall.ETH_P_ALL),
			Ifindex:  c.networkInterfaces[0].Index,
		}

		err = syscall.Sendto((int(fd)), buffer[:n], 0, to)
		if err != nil {
			log.Errorf("Error sending buffer: %s", err)
			return err
		}
	}

	return nil
}

// Run will start Canary
func (c *Canary) Start(ctx context.Context) error {
	go c.knockDetector(ctx)

	go func() {
		<-ctx.Done()
		c.Close()
	}()

	var (
		events [MaxEpollEvents]syscall.EpollEvent
		buffer [DefaultBufferSize]byte
	)

	go func() {
		log.Info("Raw listener started.")
		defer log.Info("Raw listener stopped.")

		for {
			nevents, err := syscall.EpollWait(c.epfd, events[:], -1)
			if err == nil {
			} else if errno, ok := err.(syscall.Errno); !ok {
				log.Fatalf("Error epollwait: %s", err.Error())
				return
			} else if errno.Temporary() {
				log.Errorf("Temporary epollwait error: %s, retrying.", err.Error())
				continue
			} else {
				log.Fatalf("Error epollwait: %s", err.Error())
				return
			}

			for ev := 0; ev < nevents; ev++ {
				if events[ev].Events&syscall.EPOLLIN == syscall.EPOLLIN {
					if n, _, err := syscall.Recvfrom(int(events[ev].Fd), buffer[:], 0); err != nil {
						log.Errorf("Could not receive from descriptor: %s", err.Error())
						return
					} else if n == 0 {
						// no packets received
					} else if eh, err := ethernet.Parse(buffer[:n]); err != nil {
					} else if eh.Type == EthernetTypeARP {
						data := make([]byte, len(eh.Payload))
						copy(data, eh.Payload[:])
						c.handleARP(data)
					} else if eh.Type == EthernetTypeIPv4 {
						if iph, err := ipv4.Parse(eh.Payload[:]); err != nil {
							log.Debugf("Error parsing ip header: %s", err.Error())
						} else {
							data := make([]byte, len(iph.Payload))
							copy(data, iph.Payload[:])

							switch iph.Protocol {
							case 1 /* icmp */ :
								c.handleICMP(eh, iph, data)
							case 2 /* IGMP */ :

							case 6 /* tcp */ :
								// what interface?
								c.handleTCP(eh, iph, data)
							case 17 /* udp */ :
								c.handleUDP(eh, iph, data)
							default:
								log.Debugf("Ignoring protocol: %x", iph.Protocol)
							}
						}
					}
				}

				if events[ev].Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
					c.transmit(events[ev].Fd)

					// disable epollout again
					syscall.EpollCtl(c.epfd, syscall.EPOLL_CTL_MOD, int(events[ev].Fd), &syscall.EpollEvent{
						Events: syscall.EPOLLIN,
						Fd:     int32(events[ev].Fd),
					})
				}

				if events[ev].Events&syscall.EPOLLERR == syscall.EPOLLERR {
					if v, err := syscall.GetsockoptInt(int(events[ev].Fd), syscall.SOL_SOCKET, syscall.SO_ERROR); err != nil {
						log.Errorf("Retrieving polling error: %s", err)
					} else {
						log.Errorf("Polling error: %#q", v)
					}
				}
			}
		}
	}()

	return nil
}
