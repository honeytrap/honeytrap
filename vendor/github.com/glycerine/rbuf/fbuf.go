package rbuf

// copyright (c) 2016, Jason E. Aten
// license: MIT

import (
	"io"
)

// Float64RingBuf:
//
//  a fixed-size circular ring buffer of float64
//
type Float64RingBuf struct {
	A        []float64
	N        int // MaxView, the total size of A, whether or not in use.
	Beg      int // start of in-use data in A
	Readable int // number of pointers available in A (in use)
}

// constructor. NewFloat64RingBuf will allocate internally
// a slice of size maxViewInBytes.
func NewFloat64RingBuf(maxViewInBytes int) *Float64RingBuf {
	n := maxViewInBytes
	r := &Float64RingBuf{
		N:        n,
		Beg:      0,
		Readable: 0,
	}
	r.A = make([]float64, n, n)

	return r
}

// TwoContig returns all readable pointers, but in two separate slices,
// to avoid copying. The two slices are from the same buffer, but
// are not contiguous. Either or both may be empty slices.
func (b *Float64RingBuf) TwoContig(makeCopy bool) (first []float64, second []float64) {

	extent := b.Beg + b.Readable
	if extent <= b.N {
		// we fit contiguously in this buffer without wrapping to the other.
		// Let second stay an empty slice.
		return b.A[b.Beg:(b.Beg + b.Readable)], second
	}

	return b.A[b.Beg:b.N], b.A[0:(extent % b.N)]
}

// Earliest returns the earliest written value v. ok will be
// true unless the ring is empty, in which case ok will be false,
// and v will be zero.
func (b *Float64RingBuf) Earliest() (v float64, ok bool) {
	if b.Readable == 0 {
		return
	}

	return b.A[b.Beg], true
}

// ReadFloat64():
//
// from bytes.Buffer.Read(): Read reads the next len(p) float64
// pointers from the buffer or until the buffer is drained. The return
// value n is the number of bytes read. If the buffer has no data
// to return, err is io.EOF (unless len(p) is zero); otherwise it is nil.
func (b *Float64RingBuf) ReadFloat64(p []float64) (n int, err error) {
	return b.readAndMaybeAdvance(p, true)
}

// ReadWithoutAdvance(): if you want to Read the data and leave
// it in the buffer, so as to peek ahead for example.
func (b *Float64RingBuf) ReadWithoutAdvance(p []float64) (n int, err error) {
	return b.readAndMaybeAdvance(p, false)
}

func (b *Float64RingBuf) readAndMaybeAdvance(p []float64, doAdvance bool) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if b.Readable == 0 {
		return 0, io.EOF
	}
	extent := b.Beg + b.Readable
	if extent <= b.N {
		n += copy(p, b.A[b.Beg:extent])
	} else {
		n += copy(p, b.A[b.Beg:b.N])
		if n < len(p) {
			n += copy(p[n:], b.A[0:(extent%b.N)])
		}
	}
	if doAdvance {
		b.Advance(n)
	}
	return
}

//
// WriteAndMaybeOverwriteOldestData always consumes the full
// buffer p, even if that means blowing away the oldest
// unread bytes in the ring to make room. In reality, only the last
// min(len(p),b.N) bytes of p will end up being written to the ring.
//
// This allows the ring to act as a record of the most recent
// b.N bytes of data -- a kind of temporal LRU cache, so the
// speak. The linux kernel's dmesg ring buffer is similar.
//
func (b *Float64RingBuf) WriteAndMaybeOverwriteOldestData(p []float64) (n int, err error) {
	writeCapacity := b.N - b.Readable
	if len(p) > writeCapacity {
		b.Advance(len(p) - writeCapacity)
	}
	startPos := 0
	if len(p) > b.N {
		startPos = len(p) - b.N
	}
	n, err = b.Write(p[startPos:])
	if err != nil {
		return n, err
	}
	return len(p), nil
}

//
// Write writes len(p) float64 values from p to
// the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
//
func (b *Float64RingBuf) Write(p []float64) (n int, err error) {
	for {
		if len(p) == 0 {
			// nothing (left) to copy in; notice we shorten our
			// local copy p (below) as we read from it.
			return
		}

		writeCapacity := b.N - b.Readable
		if writeCapacity <= 0 {
			// we are all full up already.
			return n, io.ErrShortWrite
		}
		if len(p) > writeCapacity {
			err = io.ErrShortWrite
			// leave err set and
			// keep going, write what we can.
		}

		writeStart := (b.Beg + b.Readable) % b.N

		upperLim := intMin(writeStart+writeCapacity, b.N)

		k := copy(b.A[writeStart:upperLim], p)

		n += k
		b.Readable += k
		p = p[k:]

		// we can fill from b.A[0:something] from
		// p's remainder, so loop
	}
}

// Reset quickly forgets any data stored in the ring buffer. The
// data is still there, but the ring buffer will ignore it and
// overwrite those buffers as new data comes in.
func (b *Float64RingBuf) Reset() {
	b.Beg = 0
	b.Readable = 0
}

// Advance(): non-standard, but better than Next(),
// because we don't have to unwrap our buffer and pay the cpu time
// for the copy that unwrapping may need.
// Useful in conjuction/after ReadWithoutAdvance() above.
func (b *Float64RingBuf) Advance(n int) {
	if n <= 0 {
		return
	}
	if n > b.Readable {
		n = b.Readable
	}
	b.Readable -= n
	b.Beg = (b.Beg + n) % b.N
}

// Adopt(): non-standard.
//
// For efficiency's sake, (possibly) take ownership of
// already allocated slice offered in me.
//
// If me is large we will adopt it, and we will potentially then
// write to the me buffer.
// If we already have a bigger buffer, copy me into the existing
// buffer instead.
func (b *Float64RingBuf) Adopt(me []float64) {
	n := len(me)
	if n > b.N {
		b.A = me
		b.N = n
		b.Beg = 0
		b.Readable = n
	} else {
		// we already have a larger buffer, reuse it.
		copy(b.A, me)
		b.Beg = 0
		b.Readable = n
	}
}
