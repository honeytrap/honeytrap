// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package test contains an integration test for all of upspin.
package test

import (
	"fmt"
	"math/rand"
	"testing"

	"upspin.io/test/testenv"
	"upspin.io/upspin"

	_ "upspin.io/dir/inprocess"
	_ "upspin.io/pack/ee"
	_ "upspin.io/pack/eeintegrity"
	_ "upspin.io/pack/plain"
	_ "upspin.io/store/inprocess"
)

func TestClientFile(t *testing.T) {
	for _, p := range []upspin.Packing{upspin.PlainPack, upspin.EEIntegrityPack, upspin.EEPack} {
		t.Run(fmt.Sprintf("packing=%v", p), func(t *testing.T) {
			testFileSequentialAccess(t, newEnv(t, p))
		})
	}
}

// newEnv configures a test environment using a packing.
func newEnv(t *testing.T, packing upspin.Packing) *testenv.Env {
	println(packing.String())
	s := &testenv.Setup{
		OwnerName: "user1@domain.com",
		Kind:      "inprocess",
		Packing:   packing,
	}
	env, err := testenv.New(s)
	if err != nil {
		t.Fatal(err)
	}
	return env
}

func setupFileIO(fileName upspin.PathName, max int, env *testenv.Env, t *testing.T) (upspin.File, []byte) {
	f, err := env.Client.Create(fileName)
	if err != nil {
		t.Fatal("create file:", err)
	}

	// Create a data set with each byte equal to its offset.
	data := make([]byte, max)
	for i := range data {
		data[i] = uint8(i)
	}
	return f, data
}

func testFileSequentialAccess(t *testing.T, env *testenv.Env) {
	client := env.Client
	userName := env.Setup.OwnerName

	const Max = 100 * 1000 // Must be > 100.
	fileName := upspin.PathName(userName + "/" + "file")
	f, data := setupFileIO(fileName, Max, env, t)

	// Write the file in randomly sized chunks until it's full.
	for offset, length := 0, 0; offset < Max; offset += length {
		// Pick a random length.
		length = rand.Intn(Max / 100)
		if offset+length > Max {
			length = Max - offset
		}
		n, err := f.Write(data[offset : offset+length])
		if err != nil {
			t.Fatalf("Write(offset %d length %d): %v", offset, length, err)
		}
		if n != length {
			t.Fatalf("Write length failed: offset %d expected %d got %d", offset, length, n)
		}
	}
	err := f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Now read it back with a similar scan.
	f, err = client.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, Max)
	for offset, length := 0, 0; offset < Max; offset += length {
		length = rand.Intn(Max / 100)
		if offset+length > Max {
			length = Max - offset
		}
		n, err := f.Read(buf[offset : offset+length])
		if err != nil {
			t.Fatalf("Read(offset %d length %d): %v", offset, length, err)
		}
		if n != length {
			t.Fatalf("Read length failed: offset %d expected %d got %d", offset, length, n)
		}
		for i := offset; i < offset+length; i++ {
			if buf[i] != data[i] {
				t.Fatalf("Read at %d (%#x): expected %#.2x got %#.2x", i, i, data[i], buf[i])
			}
		}
	}
}

func testSequenceNumbers(t *testing.T, r *testenv.Runner) {
	const (
		base   = ownerName + "/sequencenumbers"
		dir    = base + "/dir"
		subdir = dir + "/subdir"
		file   = dir + "/file"
	)
	r.As(ownerName)
	r.MakeDirectory(base)
	r.DirLookup(base)
	seq := int64(upspin.SeqBase)
	if !r.GotEntryWithSequenceVersion(base, seq) {
		t.Fatal(r.Diag())
	}

	seq++
	r.MakeDirectory(dir)
	r.DirLookup(base)
	if !r.GotEntryWithSequenceVersion(base, seq) {
		t.Fatal(r.Diag())
	}

	seq++
	r.MakeDirectory(subdir)
	r.DirLookup(base)
	if !r.GotEntryWithSequenceVersion(base, seq) {
		t.Fatal(r.Diag())
	}

	r.DirLookup(subdir)
	if !r.GotEntryWithSequenceVersion(subdir, upspin.SeqBase) {
		t.Fatal(r.Diag())
	}

	seq++
	r.Delete(subdir)
	r.DirLookup(base)
	if !r.GotEntryWithSequenceVersion(base, seq) {
		t.Fatal(r.Diag())
	}

	fileSeq := int64(upspin.SeqBase)
	r.Put(file, "meh")
	r.DirLookup(file)
	if !r.GotEntryWithSequenceVersion(file, fileSeq) {
		t.Fatal(r.Diag())
	}

	fileSeq++
	r.Put(file, "new")
	r.DirLookup(file)
	if !r.GotEntryWithSequenceVersion(file, fileSeq) {
		t.Fatal(r.Diag())
	}
}
