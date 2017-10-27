package rbuf

import (
	"fmt"
	"testing"

	cv "github.com/smartystreets/goconvey/convey"
)

func TestPointerReadWrite(t *testing.T) {
	b := NewPointerRingBuf(5)

	data := []interface{}{}
	for i := 0; i < 10; i++ {
		data = append(data, interface{}(i))
	}

	cv.Convey("PointerRingBuf::PushAndMaybeOverwriteOldestData() should auto advance", t, func() {
		b.Reset()
		n, err := b.PushAndMaybeOverwriteOldestData(data[:3])
		cv.So(err, cv.ShouldEqual, nil)
		cv.So(n, cv.ShouldEqual, 3)
		cv.So(b.Readable, cv.ShouldEqual, 3)

		n, err = b.PushAndMaybeOverwriteOldestData(data[3:5])
		cv.So(n, cv.ShouldEqual, 2)
		cv.So(b.Readable, cv.ShouldEqual, 5)
		check := make([]interface{}, 5)
		n, err = b.ReadPtrs(check)
		cv.So(n, cv.ShouldEqual, 5)
		cv.So(check, cv.ShouldResemble, data[:5])

		n, err = b.PushAndMaybeOverwriteOldestData(data[5:10])
		cv.So(err, cv.ShouldEqual, nil)
		cv.So(n, cv.ShouldEqual, 5)

		n, err = b.ReadWithoutAdvance(check)
		cv.So(n, cv.ShouldEqual, 5)
		cv.So(check, cv.ShouldResemble, data[5:10])

		// check TwoConfig
		q, r := b.TwoContig()

		//p("len q = %v", len(q))
		//p("len r = %v", len(r))

		found := make([]bool, 10)
		for _, iface := range q {
			q0 := iface.(int)
			found[q0] = true
		}

		for _, iface := range r {
			r0 := iface.(int)
			found[r0] = true
		}

		totTrue := 0
		for i := range found {
			if found[i] {
				totTrue++
			}
		}
		cv.So(totTrue, cv.ShouldEqual, 5)

	})

}

func p(format string, a ...interface{}) {
	fmt.Printf("\n"+format+"\n", a...)
}
