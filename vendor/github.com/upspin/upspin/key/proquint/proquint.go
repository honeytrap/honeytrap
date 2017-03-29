// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package proquint converts uint16 to/from pronounceable five letters.
package proquint // import "upspin.io/key/proquint"

import "bytes"

// See http://arxiv.org/html/0901.4016.
var (
	cons  = []byte("bdfghjklmnprstvz")
	vowel = []byte("aiou")
)

// Encode returns a five-letter word representing a uint16.
func Encode(x uint16) (s []byte) {
	cons3 := x & 0x0f
	x >>= 4
	vow2 := x & 0x03
	x >>= 2
	cons2 := x & 0x0f
	x >>= 4
	vow1 := x & 0x03
	x >>= 2
	cons1 := x & 0x0f
	s = make([]byte, 5)
	s[0] = cons[cons1]
	s[1] = vowel[vow1]
	s[2] = cons[cons2]
	s[3] = vowel[vow2]
	s[4] = cons[cons3]
	return
}

// Decode parses a five-letter word, returning a uint16.
func Decode(s []byte) (x uint16) {
	cons1 := uint16(bytes.IndexByte(cons, s[0]))
	vow1 := uint16(bytes.IndexByte(vowel, s[1]))
	cons2 := uint16(bytes.IndexByte(cons, s[2]))
	vow2 := uint16(bytes.IndexByte(vowel, s[3]))
	cons3 := uint16(bytes.IndexByte(cons, s[4]))
	return (((cons1<<2|vow1)<<4|cons2)<<2|vow2)<<4 | cons3
}
