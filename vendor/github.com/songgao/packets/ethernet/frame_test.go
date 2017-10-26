package ethernet

import (
	"bytes"
	"net"
	"testing"
)

func panics(f func()) (didPanic bool) {
	defer func() {
		if r := recover(); r != nil {
			didPanic = true
		}
	}()
	f()
	return
}

func mustParseMAC(str string) (addr net.HardwareAddr) {
	var err error
	addr, err = net.ParseMAC(str)
	if err != nil {
		panic(err)
	}
	return
}

func TestPrepare(t *testing.T) {
	var frame Frame
	dst := mustParseMAC("ff:ff:ff:ff:ff:ff")
	src := mustParseMAC("12:34:56:78:9a:bc")
	(&frame).Prepare(dst, src, NotTagged, IPv6, 1024)
	if len(frame.Payload()) != 1024 {
		t.Fatalf("frame payload does not have correct length. expected %d; got %d\n", 1024, len(frame.Payload()))
	}
	expectedLength := 6 + 6 + int(NotTagged) + 2 + 1024
	if len(frame) != expectedLength {
		t.Fatalf("frame does not have correct length. expected %d; got %d\n", expectedLength, len(frame))
	}
	if !bytes.Equal([]byte(frame.Source()), []byte(src)) {
		t.Fatalf("frame source address is incorrect. expected %s; got %s\n", src.String(), frame.Source().String())
	}
	if !bytes.Equal([]byte(frame.Destination()), []byte(dst)) {
		t.Fatalf("frame destination address is incorrect. expected %s; got %s\n", dst.String(), frame.Destination().String())
	}
	if frame.Tagging() != NotTagged {
		t.Fatalf("frame tagging is incorrect. expected %d; got %d\n", NotTagged, frame.Tagging())
	}
	if frame.Ethertype() != IPv6 {
		t.Fatalf("frame ethertype is incorrect. expected %v; got %v\n", IPv6, frame.Ethertype())
	}
}

func TestResize(t *testing.T) {
	var frame Frame
	(&frame).Resize(8)
	expectedLength := 6 + 6 + int(NotTagged) + 2 + 8
	if len(frame) != expectedLength {
		t.Fatalf("frame does not have correct length. expected %d; got %d\n", expectedLength, len(frame))
	}
	frame.Payload()[0] = 42
	(&frame).Resize(1024)
	if frame.Payload()[0] != 42 {
		t.Fatalf("expanded frame does not have same content\n")
	}
}
