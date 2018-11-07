// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dhcp

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/netstack/rand"
	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/udp"
	"github.com/google/netstack/waiter"
)

// Client is a DHCP client.
type Client struct {
	stack        *stack.Stack
	nicid        tcpip.NICID
	linkAddr     tcpip.LinkAddress
	acquiredFunc func(old, new tcpip.Address, cfg Config)

	mu          sync.Mutex
	addr        tcpip.Address
	cfg         Config
	lease       time.Duration
	cancelRenew func()
}

// NewClient creates a DHCP client.
//
// TODO: add s.LinkAddr(nicid) to *stack.Stack.
func NewClient(s *stack.Stack, nicid tcpip.NICID, linkAddr tcpip.LinkAddress, acquiredFunc func(old, new tcpip.Address, cfg Config)) *Client {
	return &Client{
		stack:        s,
		nicid:        nicid,
		linkAddr:     linkAddr,
		acquiredFunc: acquiredFunc,
	}
}

// Run starts the DHCP client.
// It will periodically search for an IP address using the Request method.
func (c *Client) Run(ctx context.Context) {
	go c.run(ctx)
}

func (c *Client) run(ctx context.Context) {
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.addr != "" {
			c.stack.RemoveAddress(c.nicid, c.addr)
		}
	}()

	var renewAddr tcpip.Address
	for {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		cfg, err := c.Request(reqCtx, renewAddr)
		cancel()
		if err != nil {
			select {
			case <-time.After(1 * time.Second):
				// loop and try again
			case <-ctx.Done():
				return
			}
		}

		c.mu.Lock()
		renewAddr = c.addr
		c.mu.Unlock()

		timer := time.NewTimer(cfg.LeaseLength)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// loop and make a renewal request
		}
	}
}

// Address reports the IP address acquired by the DHCP client.
func (c *Client) Address() tcpip.Address {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addr
}

// Config reports the DHCP configuration acquired with the IP address lease.
func (c *Client) Config() Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg
}

// Request executes a DHCP request session.
//
// On success, it adds a new address to this client's TCPIP stack.
// If the server sets a lease limit a timer is set to automatically
// renew it.
func (c *Client) Request(ctx context.Context, requestedAddr tcpip.Address) (cfg Config, reterr error) {
	if err := c.stack.AddAddressWithOptions(c.nicid, ipv4.ProtocolNumber, "\xff\xff\xff\xff", stack.NeverPrimaryEndpoint); err != nil && err != tcpip.ErrDuplicateAddress {
		return Config{}, fmt.Errorf("dhcp: %v", err)
	}
	if err := c.stack.AddAddressWithOptions(c.nicid, ipv4.ProtocolNumber, "\x00\x00\x00\x00", stack.NeverPrimaryEndpoint); err != nil && err != tcpip.ErrDuplicateAddress {
		return Config{}, fmt.Errorf("dhcp: %v", err)
	}
	defer c.stack.RemoveAddress(c.nicid, "\xff\xff\xff\xff")
	defer c.stack.RemoveAddress(c.nicid, "\x00\x00\x00\x00")

	var wq waiter.Queue
	ep, err := c.stack.NewEndpoint(udp.ProtocolNumber, ipv4.ProtocolNumber, &wq)
	if err != nil {
		return Config{}, fmt.Errorf("dhcp: outbound endpoint: %v", err)
	}
	defer ep.Close()
	if err := ep.Bind(tcpip.FullAddress{
		Addr: "\x00\x00\x00\x00",
		Port: ClientPort,
		NIC:  c.nicid,
	}, nil); err != nil {
		return Config{}, fmt.Errorf("dhcp: connect failed: %v", err)
	}

	epin, err := c.stack.NewEndpoint(udp.ProtocolNumber, ipv4.ProtocolNumber, &wq)
	if err != nil {
		return Config{}, fmt.Errorf("dhcp: inbound endpoint: %v", err)
	}
	defer epin.Close()
	if err := epin.Bind(tcpip.FullAddress{
		Addr: "\xff\xff\xff\xff",
		Port: ClientPort,
		NIC:  c.nicid,
	}, nil); err != nil {
		return Config{}, fmt.Errorf("dhcp: connect failed: %v", err)
	}

	var xid [4]byte
	rand.Read(xid[:])

	// DHCPDISCOVERY
	discOpts := options{
		{optDHCPMsgType, []byte{byte(dhcpDISCOVER)}},
		{optParamReq, []byte{
			1,  // request subnet mask
			3,  // request router
			15, // domain name
			6,  // domain name server
		}},
	}
	if requestedAddr != "" {
		discOpts = append(discOpts, option{optReqIPAddr, []byte(requestedAddr)})
	}
	var clientID []byte
	if len(c.linkAddr) == 6 {
		clientID = append(
			[]byte{1}, // RFC 1700: Hardware Type [Ethernet = 1]
			c.linkAddr...,
		)
		discOpts = append(discOpts, option{optClientID, clientID})
	}
	h := make(header, headerBaseSize+discOpts.len()+1)
	h.init()
	h.setOp(opRequest)
	copy(h.xidbytes(), xid[:])
	h.setBroadcast()
	copy(h.chaddr(), c.linkAddr)
	h.setOptions(discOpts)

	serverAddr := &tcpip.FullAddress{
		Addr: "\xff\xff\xff\xff",
		Port: ServerPort,
		NIC:  c.nicid,
	}
	wopts := tcpip.WriteOptions{
		To: serverAddr,
	}
	var resCh <-chan struct{}
	if _, resCh, err = ep.Write(tcpip.SlicePayload(h), wopts); err != nil && resCh == nil {
		return Config{}, fmt.Errorf("dhcp discovery write: %v", err)
	}

	if resCh != nil {
		select {
		case <-resCh:
		case <-ctx.Done():
			return Config{}, fmt.Errorf("dhcp client address resolution: %v", tcpip.ErrAborted)
		}

		if _, _, err := ep.Write(tcpip.SlicePayload(h), wopts); err != nil {
			return Config{}, fmt.Errorf("dhcp discovery write: %v", err)
		}
	}

	we, ch := waiter.NewChannelEntry(nil)
	wq.EventRegister(&we, waiter.EventIn)
	defer wq.EventUnregister(&we)

	// DHCPOFFER
	var opts options
	for {
		var addr tcpip.FullAddress
		v, _, err := epin.Read(&addr)
		if err == tcpip.ErrWouldBlock {
			select {
			case <-ch:
				continue
			case <-ctx.Done():
				return Config{}, fmt.Errorf("reading dhcp offer: %v", tcpip.ErrAborted)
			}
		}
		h = header(v)
		var valid bool
		var e error
		opts, valid, e = loadDHCPReply(h, dhcpOFFER, xid[:])
		if !valid {
			if e != nil {
				// TODO: handle all the errors?
				// TODO: report malformed server responses
			}
			continue
		}
		break
	}

	var ack bool
	if err := cfg.decode(opts); err != nil {
		return Config{}, fmt.Errorf("dhcp offer: %v", err)
	}

	// DHCPREQUEST
	addr := tcpip.Address(h.yiaddr())
	if err := c.stack.AddAddressWithOptions(c.nicid, ipv4.ProtocolNumber, addr, stack.FirstPrimaryEndpoint); err != nil {
		if err != tcpip.ErrDuplicateAddress {
			return Config{}, fmt.Errorf("adding address: %v", err)
		}
	}
	defer func() {
		if !ack || reterr != nil {
			c.stack.RemoveAddress(c.nicid, addr)
			addr = ""
			cfg = Config{Error: reterr}
		}

		c.mu.Lock()
		oldAddr := c.addr
		c.addr = addr
		c.cfg = cfg
		c.mu.Unlock()

		// Clean up broadcast addresses before calling acquiredFunc
		// so nothing else uses them by mistake.
		//
		// (The deferred RemoveAddress calls above silently error.)
		c.stack.RemoveAddress(c.nicid, "\xff\xff\xff\xff")
		c.stack.RemoveAddress(c.nicid, "\x00\x00\x00\x00")

		if c.acquiredFunc != nil {
			c.acquiredFunc(oldAddr, addr, cfg)
		}
		if requestedAddr != "" && requestedAddr != addr {
			c.stack.RemoveAddress(c.nicid, requestedAddr)
		}
	}()
	h.init()
	h.setOp(opRequest)
	for i, b := 0, h.yiaddr(); i < len(b); i++ {
		b[i] = 0
	}
	for i, b := 0, h.siaddr(); i < len(b); i++ {
		b[i] = 0
	}
	for i, b := 0, h.giaddr(); i < len(b); i++ {
		b[i] = 0
	}
	reqOpts := []option{
		{optDHCPMsgType, []byte{byte(dhcpREQUEST)}},
		{optReqIPAddr, []byte(addr)},
		{optDHCPServer, []byte(cfg.ServerAddress)},
	}
	if len(clientID) != 0 {
		reqOpts = append(reqOpts, option{optClientID, clientID})
	}
	h.setOptions(reqOpts)
	if _, _, err := ep.Write(tcpip.SlicePayload(h), wopts); err != nil {
		return Config{}, fmt.Errorf("dhcp discovery write: %v", err)
	}

	// DHCPACK
	for {
		var addr tcpip.FullAddress
		v, _, err := epin.Read(&addr)
		if err == tcpip.ErrWouldBlock {
			select {
			case <-ch:
				continue
			case <-ctx.Done():
				return Config{}, fmt.Errorf("reading dhcp ack: %v", tcpip.ErrAborted)
			}
		}
		h = header(v)
		var valid bool
		var e error
		opts, valid, e = loadDHCPReply(h, dhcpACK, xid[:])
		if !valid {
			if e != nil {
				// TODO: handle all the errors?
				// TODO: report malformed server responses
			}
			if opts, valid, _ = loadDHCPReply(h, dhcpNAK, xid[:]); valid {
				if msg := opts.message(); msg != "" {
					return Config{}, fmt.Errorf("dhcp: NAK %q", msg)
				}
				return Config{}, fmt.Errorf("dhcp: NAK with no message")
			}
			continue
		}
		break
	}
	ack = true
	return cfg, nil
}

func loadDHCPReply(h header, typ dhcpMsgType, xid []byte) (opts options, valid bool, err error) {
	if !h.isValid() || h.op() != opReply || !bytes.Equal(h.xidbytes(), xid[:]) {
		return nil, false, nil
	}
	opts, err = h.options()
	if err != nil {
		return nil, false, err
	}
	msgtype, err := opts.dhcpMsgType()
	if err != nil {
		return nil, false, err
	}
	if msgtype != typ {
		return nil, false, nil
	}
	return opts, true, nil
}
