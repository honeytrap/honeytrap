package ethernet

import "net"

// Frame represents an ethernet frame. The length of the underlying slice of a
// Frame should always reflect the ethernet frame length.
type Frame []byte

// Tagging is a type used to indicate whether/how a frame is tagged. The value
// is number of bytes taken by tagging.
type Tagging byte

// Const values for different taggings
const (
	NotTagged    Tagging = 0
	Tagged       Tagging = 4
	DoubleTagged Tagging = 8
)

// Destination returns the destination address field of the frame. The address
// references a slice on the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Destination() net.HardwareAddr {
	return net.HardwareAddr(f[:6:6])
}

// Source returns the source address field of the frame. The address references
// a slice on the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Source() net.HardwareAddr {
	return net.HardwareAddr(f[6:12:12])
}

// Tagging returns whether/how the frame has 802.1Q tag(s).
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Tagging() Tagging {
	if f[12] == 0x81 && f[13] == 0x00 {
		return Tagged
	} else if f[12] == 0x88 && f[13] == 0xa8 {
		return DoubleTagged
	}
	return NotTagged
}

// Tag returns a slice holding the tag part of the frame, if any. Note that
// this includes the Tag Protocol Identifier (TPID), e.g. 0x8100 or 0x88a8.
// Upper layer should use the returned slice for both reading and writing.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Tags() []byte {
	tagging := f.Tagging()
	return f[12 : 12+tagging : 12+tagging]
}

// Ethertype returns the ethertype field of the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Ethertype() Ethertype {
	ethertypePos := 12 + f.Tagging()
	return Ethertype{f[ethertypePos], f[ethertypePos+1]}
}

// Payload returns a slice holding the payload part of the frame. Upper layer
// should use the returned slice for both reading and writing purposes.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f Frame) Payload() []byte {
	return f[12+f.Tagging()+2:]
}

// Resize re-slices (*f) so that len(*f) holds exactly payloadSize bytes of
// payload. If cap(*f) is not large enough, a new slice is made and content
// from old slice is copied to the new one.
//
// If len(*f) is less than 14 bytes, it is assumed to be not tagged.
//
// It is safe to call Resize on a pointer to a nil Frame.
func (f *Frame) Resize(payloadSize int) {
	tagging := NotTagged
	if len(*f) > 6+6+2 {
		tagging = f.Tagging()
	}
	f.resize(6 + 6 + int(tagging) + 2 + payloadSize)
}

// Prepare prepares *f to be used, by filling in dst/src address, setting up
// proper tagging and ethertype, and resizing it to proper length.
//
// It is safe to call Prepare on a pointer to a nil Frame or invalid Frame.
func (f *Frame) Prepare(dst net.HardwareAddr, src net.HardwareAddr, tagging Tagging, ethertype Ethertype, payloadSize int) {
	f.resize(6 + 6 + int(tagging) + 2 + payloadSize)
	copy((*f)[0:6:6], dst)
	copy((*f)[6:12:12], src)
	if tagging == Tagged {
		(*f)[12] = 0x81
		(*f)[13] = 0x00
	} else if tagging == DoubleTagged {
		(*f)[12] = 0x88
		(*f)[13] = 0xa8
	}
	(*f)[12+tagging] = ethertype[0]
	(*f)[12+tagging+1] = ethertype[1]
	return
}

func (f *Frame) resize(length int) {
	if cap(*f) < length {
		old := *f
		*f = make(Frame, length, length)
		copy(*f, old)
	} else {
		*f = (*f)[:length]
	}
}
