/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// toy VNC (RFB) server in Go, just learning the protocol.
//
// Protocol docs:
//    http://www.realvnc.com/docs/rfbproto.pdf
//
// Author: Brad Fitzpatrick <brad@danga.com>

package vnc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"image"
	"net"
	"sync"
)

const (
	v3 = "RFB 003.003\n"
	v7 = "RFB 003.007\n"
	v8 = "RFB 003.008\n"

	authNone = 1

	statusOK     = 0
	statusFailed = 1

	encodingRaw      = 0
	encodingCopyRect = 1

	// Client -> Server
	cmdSetPixelFormat           = 0
	cmdSetEncodings             = 2
	cmdFramebufferUpdateRequest = 3
	cmdKeyEvent                 = 4
	cmdPointerEvent             = 5
	cmdClientCutText            = 6

	// Server -> Client
	cmdFramebufferUpdate = 0
)

func newConn(width, height int, c net.Conn) *Conn {
	feed := make(chan *LockableImage, 16)
	event := make(chan interface{}, 16)
	conn := &Conn{
		height:     height,
		width:      width,
		c:          c,
		serverName: "",
		br:         bufio.NewReader(c),
		bw:         bufio.NewWriter(c),
		fbupc:      make(chan FrameBufferUpdateRequest, 128),
		closec:     make(chan bool),
		feed:       feed,
		Feed:       feed, // the send-only version
		event:      event,
		Event:      event, // the recieve-only version
	}
	return conn
}

type LockableImage struct {
	sync.RWMutex
	Img image.Image
}

type Conn struct {
	serverName string

	c      net.Conn
	br     *bufio.Reader
	bw     *bufio.Writer
	fbupc  chan FrameBufferUpdateRequest
	closec chan bool // never sent; just closed

	// should only be mutated once during handshake, but then
	// only read.
	format PixelFormat

	feed chan *LockableImage
	mu   sync.RWMutex // guards last (but not its pixels, just the variable)
	last *LockableImage

	buf8 []uint8 // temporary buffer to avoid generating garbage

	// Feed is the channel to send new frames.
	Feed chan<- *LockableImage

	// Event is a readable channel of events from the client.
	// The value will be either a KeyEvent or PointerEvent.  The
	// channel is closed when the client disconnects.
	Event <-chan interface{}

	event chan interface{} // internal version of Event

	height int
	width  int

	gotFirstFrame bool
}

func (c *Conn) dimensions() (w, h int) {
	return c.width, c.height
}

func (c *Conn) readByte(what string) byte {
	b, err := c.br.ReadByte()
	if err != nil {
		c.failf("reading client byte for %q: %v", what, err)
	}
	return b
}

func (c *Conn) readPadding(what string, size int) {
	for i := 0; i < size; i++ {
		c.readByte(what)
	}
}

func (c *Conn) read(what string, v interface{}) {
	err := binary.Read(c.br, binary.BigEndian, v)
	if err != nil {
		c.failf("reading from client into %T for %q: %v", v, what, err)
	}
}

func (c *Conn) w(v interface{}) {
	binary.Write(c.bw, binary.BigEndian, v)
}

func (c *Conn) flush() {
	c.bw.Flush()
}

func (c *Conn) failf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (c *Conn) serve() {
	defer c.c.Close()
	defer close(c.fbupc)
	defer close(c.closec)
	defer close(c.event)
	defer func() {
		e := recover()
		if e != nil {
			log.Debugf("Client disconnect: %v", e)
		}
	}()

	c.bw.WriteString("RFB 003.008\n")
	c.flush()
	sl, err := c.br.ReadSlice('\n')
	if err != nil {
		c.failf("reading client protocol version: %v", err)
	}
	ver := string(sl)
	log.Debugf("client wants: %q", ver)
	switch ver {
	case v3, v7, v8: // cool.
	default:
		c.failf("bogus client-requested security type %q", ver)
	}

	// Auth
	if ver >= v7 {
		// Just 1 auth type supported: 1 (no auth)
		c.bw.WriteString("\x01\x01")
		c.flush()
		wanted := c.readByte("6.1.2:client requested security-type")
		if wanted != authNone {
			c.failf("client wanted auth type %d, not None", int(wanted))
		}
	} else {
		// Old way. Just tell client we're doing no auth.
		c.w(uint32(authNone))
		c.flush()
	}

	if ver >= v8 {
		// 6.1.3. SecurityResult
		c.w(uint32(statusOK))
		c.flush()
	}

	log.Debugf("reading client init")

	// ClientInit
	wantShared := c.readByte("shared-flag") != 0
	_ = wantShared

	c.format = PixelFormat{
		BPP:        16,
		Depth:      16,
		BigEndian:  0,
		TrueColour: 1,
		RedMax:     0x1f,
		GreenMax:   0x1f,
		BlueMax:    0x1f,
		RedShift:   0xa,
		GreenShift: 0x5,
		BlueShift:  0,
	}

	// 6.3.2. ServerInit
	width, height := c.dimensions()
	c.w(uint16(width))
	c.w(uint16(height))
	c.w(c.format.BPP)
	c.w(c.format.Depth)
	c.w(c.format.BigEndian)
	c.w(c.format.TrueColour)
	c.w(c.format.RedMax)
	c.w(c.format.GreenMax)
	c.w(c.format.BlueMax)
	c.w(c.format.RedShift)
	c.w(c.format.GreenShift)
	c.w(c.format.BlueShift)
	c.w(uint8(0)) // pad1
	c.w(uint8(0)) // pad2
	c.w(uint8(0)) // pad3
	c.w(int32(len(c.serverName)))
	c.bw.WriteString(c.serverName)
	c.flush()

	for {
		//log.Debugf("awaiting command byte from client...")
		cmd := c.readByte("6.4:client-server-packet-type")
		//log.Debugf("got command type %d from client", int(cmd))
		switch cmd {
		case cmdSetPixelFormat:
			c.handleSetPixelFormat()
		case cmdSetEncodings:
			c.handleSetEncodings()
		case cmdFramebufferUpdateRequest:
			c.handleUpdateRequest()
		case cmdPointerEvent:
			c.handlePointerEvent()
		case cmdKeyEvent:
			c.handleKeyEvent()
		default:
			c.failf("unsupported command type %d from client", int(cmd))
		}
	}
}

func (c *Conn) pushFramesLoop() {
	for {
		select {
		case ur, ok := <-c.fbupc:
			if !ok {
				// Client disconnected.
				return
			}
			c.pushFrame(ur)
		case li := <-c.feed:
			c.mu.Lock()
			c.last = li
			c.mu.Unlock()
			c.pushImage(li)
		}
	}
}

func (c *Conn) pushFrame(ur FrameBufferUpdateRequest) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	li := c.last
	if li == nil {
		return
	}

	if ur.incremental() {
		li.Lock()
		defer li.Unlock()
		im := li.Img
		b := im.Bounds()
		width, height := b.Dx(), b.Dy()

		//log.Debugf("Client wants incremental update, sending none. %#v", ur)
		c.w(uint8(cmdFramebufferUpdate))
		c.w(uint8(0))      // padding byte
		c.w(uint16(1))     // no rectangles
		c.w(uint16(0))     // x
		c.w(uint16(0))     // y
		c.w(uint16(width)) // x
		c.w(uint16(height))
		c.w(int32(encodingCopyRect))
		c.w(uint16(0)) // src-x
		c.w(uint16(0)) // src-y
		c.flush()
		return
	}
	c.pushImage(li)
}

func (c *Conn) pushImage(li *LockableImage) {
	li.Lock()
	defer li.Unlock()

	im := li.Img
	b := im.Bounds()
	if b.Min.X != 0 || b.Min.Y != 0 {
		log.Errorf("this code is lazy and assumes images with Min bounds at 0,0")
		return
	}
	width, height := b.Dx(), b.Dy()

	c.w(uint8(cmdFramebufferUpdate))
	c.w(uint8(0))  // padding byte
	c.w(uint16(1)) // 1 rectangle

	//log.Debugf("sending %d x %d pixels", width, height)

	if c.format.TrueColour == 0 {
		c.failf("only true-colour supported")
	}

	// Send that rectangle:
	c.w(uint16(0))     // x
	c.w(uint16(0))     // y
	c.w(uint16(width)) // x
	c.w(uint16(height))
	c.w(int32(encodingRaw))

	rgba, isRGBA := im.(*image.RGBA)
	if isRGBA && c.format.isScreensThousands() {
		// Fast path.
		c.pushRGBAScreensThousandsLocked(rgba)
	} else {
		c.pushGenericLocked(im)
	}
	c.flush()
}

func (c *Conn) pushRGBAScreensThousandsLocked(im *image.RGBA) {
	var u16 uint16
	pixels := len(im.Pix) / 4
	if len(c.buf8) < pixels*2 {
		c.buf8 = make([]byte, pixels*2)
	}
	out := c.buf8[:]
	isBigEndian := c.format.BigEndian != 0
	for i, v8 := range im.Pix {
		switch i % 4 {
		case 0: // red
			u16 = uint16(v8&248) << 7 // 3 masked bits + 7 shifted == redshift of 10
		case 1: // green
			u16 |= uint16(v8&248) << 2 // redshift of 5
		case 2: // blue
			u16 |= uint16(v8 >> 3)
		case 3: // alpha, unused.  use this to just move the dest
			hb, lb := uint8(u16>>8), uint8(u16)
			if isBigEndian {
				out[0] = hb
				out[1] = lb
			} else {
				out[0] = lb
				out[1] = hb
			}
			out = out[2:]
		}
	}
	c.bw.Write(c.buf8[:pixels*2])
}

// pushGenericLocked is the slow path generic implementation that works on
// any image.Image concrete type and any client-requested pixel format.
// If you're lucky, you never end in this path.
func (c *Conn) pushGenericLocked(im image.Image) {
	b := im.Bounds()
	width, height := b.Dx(), b.Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			col := im.At(x, y)
			r16, g16, b16, _ := col.RGBA()
			r16 = inRange(r16, c.format.RedMax)
			g16 = inRange(g16, c.format.GreenMax)
			b16 = inRange(b16, c.format.BlueMax)
			var u32 uint32 = (r16 << c.format.RedShift) |
				(g16 << c.format.GreenShift) |
				(b16 << c.format.BlueShift)
			var v interface{}
			switch c.format.BPP {
			case 32:
				v = u32
			case 16:
				v = uint16(u32)
			case 8:
				v = uint8(u32)
			default:
				c.failf("TODO: BPP of %d", c.format.BPP)
			}
			if c.format.BigEndian != 0 {
				binary.Write(c.bw, binary.BigEndian, v)
			} else {
				binary.Write(c.bw, binary.LittleEndian, v)
			}
		}
	}
}

type PixelFormat struct {
	BPP, Depth                      uint8
	BigEndian, TrueColour           uint8 // flags; 0 or non-zero
	RedMax, GreenMax, BlueMax       uint16
	RedShift, GreenShift, BlueShift uint8
}

// Is the format requested by the OS X "Screens" app's "Thousands" mode.
func (f *PixelFormat) isScreensThousands() bool {
	// Note: Screens asks for Depth 16; RealVNC asks for Depth 15 (which is more accurate)
	// Accept either. Same format.
	return f.BPP == 16 && (f.Depth == 16 || f.Depth == 15) && f.TrueColour != 0 &&
		f.RedMax == 0x1f && f.GreenMax == 0x1f && f.BlueMax == 0x1f &&
		f.RedShift == 10 && f.GreenShift == 5 && f.BlueShift == 0
}

// 6.4.1
func (c *Conn) handleSetPixelFormat() {
	log.Debugf("handling setpixel format")
	c.readPadding("SetPixelFormat padding", 3)
	var pf PixelFormat
	c.read("pixelformat.bpp", &pf.BPP)
	c.read("pixelformat.depth", &pf.Depth)
	c.read("pixelformat.beflag", &pf.BigEndian)
	c.read("pixelformat.truecolour", &pf.TrueColour)
	c.read("pixelformat.redmax", &pf.RedMax)
	c.read("pixelformat.greenmax", &pf.GreenMax)
	c.read("pixelformat.bluemax", &pf.BlueMax)
	c.read("pixelformat.redshift", &pf.RedShift)
	c.read("pixelformat.greenshift", &pf.GreenShift)
	c.read("pixelformat.blueshift", &pf.BlueShift)
	c.readPadding("SetPixelFormat pixel format padding", 3)
	log.Debugf("Client wants pixel format: %#v", pf)
	c.format = pf

	// TODO: send PixelFormat event? would clients care?
}

// 6.4.2
func (c *Conn) handleSetEncodings() {
	c.readPadding("SetEncodings padding", 1)

	var numEncodings uint16
	c.read("6.4.2:number-of-encodings", &numEncodings)
	var encType []int32
	for i := 0; i < int(numEncodings); i++ {
		var t int32
		c.read("encoding-type", &t)
		encType = append(encType, t)
	}
	log.Debugf("Client encodings: %#v", encType)

}

// 6.4.3
type FrameBufferUpdateRequest struct {
	IncrementalFlag     uint8
	X, Y, Width, Height uint16
}

func (r *FrameBufferUpdateRequest) incremental() bool { return r.IncrementalFlag != 0 }

// 6.4.3
func (c *Conn) handleUpdateRequest() {
	if !c.gotFirstFrame {
		li := <-c.feed
		c.mu.Lock()
		c.last = li
		c.mu.Unlock()
		c.gotFirstFrame = true
		go c.pushFramesLoop()
	}

	var req FrameBufferUpdateRequest
	c.read("framebuffer-update.incremental", &req.IncrementalFlag)
	c.read("framebuffer-update.x", &req.X)
	c.read("framebuffer-update.y", &req.Y)
	c.read("framebuffer-update.width", &req.Width)
	c.read("framebuffer-update.height", &req.Height)
	c.fbupc <- req
}

// 6.4.4
type KeyEvent struct {
	DownFlag uint8
	Key      uint32
}

// 6.4.4
func (c *Conn) handleKeyEvent() {
	var req KeyEvent
	c.read("key-event.downflag", &req.DownFlag)
	c.readPadding("key-event.padding", 2)
	c.read("key-event.key", &req.Key)
	select {
	case c.event <- req:
	default:
		// Client's too slow.
	}
}

// 6.4.5
type PointerEvent struct {
	ButtonMask uint8
	X, Y       uint16
}

// 6.4.5
func (c *Conn) handlePointerEvent() {
	var req PointerEvent
	c.read("pointer-event.mask", &req.ButtonMask)
	c.read("pointer-event.x", &req.X)
	c.read("pointer-event.y", &req.Y)
	select {
	case c.event <- req:
	default:
		// Client's too slow.
	}
}

func inRange(v uint32, max uint16) uint32 {
	switch max {
	case 0x1f: // 5 bits
		return v >> (16 - 5)
	}

	log.Errorf("unsupported inRange: v=%d, max=%d", v, max)
	return 0
}
