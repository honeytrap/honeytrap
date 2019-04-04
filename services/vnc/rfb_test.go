// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
