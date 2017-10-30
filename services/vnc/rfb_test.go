package vnc_test

import (
	"testing"
)

func BenchmarkAppend(b *testing.B) {
	var buf []uint16
	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		for v := uint16(0); v < 65535; v++ {
			buf = append(buf, v)
		}
	}
}

func BenchmarkArray(b *testing.B) {
	var buf []uint16 = make([]uint16, 65536)
	for i := 0; i < b.N; i++ {
		for v := uint16(0); v < 65535; v++ {
			buf[v] = v
		}
	}
}

func BenchmarkMod4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			_ = i
		case 1:
			_ = i
		case 2:
			_ = i
		case 3:
			_ = i
		}
	}
}

func BenchmarkMask3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		switch i & 3 {
		case 0:
			_ = i
		case 1:
			_ = i
		case 2:
			_ = i
		case 3:
			_ = i
		}
	}
}
