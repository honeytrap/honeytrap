package rbuf

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestRingBufReadWrite(t *testing.T) {
	b := NewFixedSizeRingBuf(5)

	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	cv.Convey("Given a FixedSizeRingBuf of size 5", t, func() {
		cv.Convey("Write(), Bytes(), and Read() should put and get bytes", func() {
			n, err := b.Write(data[0:5])
			cv.So(n, cv.ShouldEqual, 5)
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(b.Readable, cv.ShouldEqual, 5)
			if n != 5 {
				fmt.Printf("should have been able to write 5 bytes.\n")
			}
			if err != nil {
				panic(err)
			}
			cv.So(b.Bytes(), cv.ShouldResemble, data[0:5])

			sink := make([]byte, 3)
			n, err = b.Read(sink)
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(b.Bytes(), cv.ShouldResemble, data[3:5])
			cv.So(sink, cv.ShouldResemble, data[0:3])
		})

		cv.Convey("Write() more than 5 should give back ErrShortWrite", func() {
			b.Reset()
			cv.So(b.Readable, cv.ShouldEqual, 0)
			n, err := b.Write(data[0:10])
			cv.So(n, cv.ShouldEqual, 5)
			cv.So(err, cv.ShouldEqual, io.ErrShortWrite)
			cv.So(b.Readable, cv.ShouldEqual, 5)
			if n != 5 {
				fmt.Printf("should have been able to write 5 bytes.\n")
			}
			cv.So(b.Bytes(), cv.ShouldResemble, data[0:5])

			sink := make([]byte, 3)
			n, err = b.Read(sink)
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(b.Bytes(), cv.ShouldResemble, data[3:5])
			cv.So(sink, cv.ShouldResemble, data[0:3])
		})

		cv.Convey("we should be able to wrap data and then get it back in Bytes()", func() {
			b.Reset()

			n, err := b.Write(data[0:3])
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(err, cv.ShouldEqual, nil)

			sink := make([]byte, 3)
			n, err = b.Read(sink) // put b.beg at 3
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(b.Readable, cv.ShouldEqual, 0)

			n, err = b.Write(data[3:8]) // wrap 3 bytes around to the front
			cv.So(n, cv.ShouldEqual, 5)
			cv.So(err, cv.ShouldEqual, nil)

			by := b.Bytes()
			cv.So(by, cv.ShouldResemble, data[3:8]) // but still get them back from the ping-pong buffering

		})

		cv.Convey("FixedSizeRingBuf::WriteTo() should work with wrapped data", func() {
			b.Reset()

			n, err := b.Write(data[0:3])
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(err, cv.ShouldEqual, nil)

			sink := make([]byte, 3)
			n, err = b.Read(sink) // put b.beg at 3
			cv.So(n, cv.ShouldEqual, 3)
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(b.Readable, cv.ShouldEqual, 0)

			n, err = b.Write(data[3:8]) // wrap 3 bytes around to the front

			var bb bytes.Buffer
			m, err := b.WriteTo(&bb)

			cv.So(m, cv.ShouldEqual, 5)
			cv.So(err, cv.ShouldEqual, nil)

			by := bb.Bytes()
			cv.So(by, cv.ShouldResemble, data[3:8]) // but still get them back from the ping-pong buffering

		})

		cv.Convey("FixedSizeRingBuf::ReadFrom() should work with wrapped data", func() {
			b.Reset()
			var bb bytes.Buffer
			n, err := b.ReadFrom(&bb)
			cv.So(n, cv.ShouldEqual, 0)
			cv.So(err, cv.ShouldEqual, nil)

			// write 4, then read 4 bytes
			m, err := b.Write(data[0:4])
			cv.So(m, cv.ShouldEqual, 4)
			cv.So(err, cv.ShouldEqual, nil)

			sink := make([]byte, 4)
			k, err := b.Read(sink) // put b.beg at 4
			cv.So(k, cv.ShouldEqual, 4)
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(b.Readable, cv.ShouldEqual, 0)
			cv.So(b.Beg, cv.ShouldEqual, 4)

			bbread := bytes.NewBuffer(data[4:9])
			n, err = b.ReadFrom(bbread) // wrap 4 bytes around to the front, 5 bytes total.

			by := b.Bytes()
			cv.So(by, cv.ShouldResemble, data[4:9]) // but still get them back continguous from the ping-pong buffering
		})
		cv.Convey("FixedSizeRingBuf::WriteAndMaybeOverwriteOldestData() should auto advance", func() {
			b.Reset()
			n, err := b.WriteAndMaybeOverwriteOldestData(data[:5])
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(n, cv.ShouldEqual, 5)

			n, err = b.WriteAndMaybeOverwriteOldestData(data[5:7])
			cv.So(n, cv.ShouldEqual, 2)
			cv.So(b.Bytes(), cv.ShouldResemble, data[2:7])

			n, err = b.WriteAndMaybeOverwriteOldestData(data[0:9])
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(n, cv.ShouldEqual, 9)
			cv.So(b.Bytes(), cv.ShouldResemble, data[4:9])
		})

	})

}

func TestNextPrev(t *testing.T) {
	b := NewFixedSizeRingBuf(6)
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	cv.Convey("Given a FixedSizeRingBuf of size 6, filled with 4 elements at various begin points, then Nextpos() and Prev() should return the correct positions or <0 if done", t, func() {
		k := b.N
		for i := 0; i < b.N; i++ {
			b.Reset()
			b.Beg = i
			_, err := b.Write(data[0:k])
			panicOn(err)
			// cannot go prev to first
			cv.So(b.Prevpos(i), cv.ShouldEqual, -1)
			// cannot go after last
			cv.So(b.Nextpos((i+k-1)%b.N), cv.ShouldEqual, -1)
			// in the middle we should be okay
			for j := 1; j < k-1; j++ {
				r := (i + j) % b.N
				prev := b.Prevpos(r)
				next := b.Nextpos(r)
				cv.So(prev >= 0, cv.ShouldBeTrue)
				cv.So(next >= 0, cv.ShouldBeTrue)
				if next > r {
					cv.So(next, cv.ShouldEqual, r+1)
				} else {
					cv.So(next, cv.ShouldEqual, 0)
				}
				if prev < r {
					cv.So(prev, cv.ShouldEqual, r-1)
				} else {
					cv.So(prev, cv.ShouldEqual, b.N-1)
				}
			}
		}
	})
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
