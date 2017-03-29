// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Round 1.
const (
	secretStr = "pibud-sijat-ponam-zizaz.kudol-visin-vakok-jinok"
	publicKey = `p256
20475414006091125411730282854763965332579614918776190347990649355528840360162
41618798560597642013440926161855187887081385971895806061707694318148863738083
`

	privateKey = "103735382135370212717736500933863354513183407328603457343387144070761075604179\n"
)

// Round 2.
const (
	secretStr2 = "ponam-sijat-pibud-zizaz.vakok-visin-kudol-jinok"

	public2Key = `p256
12123951731103713463684549996372980607529899544492565862335281576402253981962
115773437512973772745017348056058736133905153504498008582636958858837556112862
`

	private2Key = "47513031334211958720530809945101319621000940818220052385943490959145831252616\n"

	archive2Key = `# EE
p256
20475414006091125411730282854763965332579614918776190347990649355528840360162
41618798560597642013440926161855187887081385971895806061707694318148863738083
103735382135370212717736500933863354513183407328603457343387144070761075604179
`
)

func TestSaveKeygen(t *testing.T) {
	public, private, _, err := createKeys("p256", secretStr)
	if err != nil {
		t.Fatalf("creating keys: %v", err)
	}
	if public != publicKey {
		t.Errorf("round 1: got public key %q; want %q", public, publicKey)
	}
	if private != privateKey {
		t.Errorf("round 1: got private key %q; want %q", private, privateKey)
	}

	// Write them to a file.
	dir, err := ioutil.TempDir("", "keygen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	err = writeKeys(dir, public, private)
	if err != nil {
		t.Fatalf("writing keys: %v", err)
	}

	// Read them back.
	data, err := ioutil.ReadFile(filepath.Join(dir, "public.upspinkey"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != public {
		t.Fatalf("reading public key: got %q; want %q", data, public)
	}
	data, err = ioutil.ReadFile(filepath.Join(dir, "secret.upspinkey"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != private {
		t.Fatalf("reading secret key: got %q; want %q", data, private)
	}

	// Generate again.
	public, private, _, err = createKeys("p256", secretStr2)
	if err != nil {
		t.Fatalf("creating keys: %v", err)
	}
	if public != public2Key {
		t.Errorf("round 2: got public key %q; want %q", public, publicKey)
	}
	if private != private2Key {
		t.Errorf("round 2: got private key %q; want %q", private, private2Key)
	}

	// Save, then write them to a file again.
	err = saveKeys(dir)
	if err != nil {
		t.Fatalf("saving keys: %v", err)
	}
	err = writeKeys(dir, public, private)
	if err != nil {
		t.Fatalf("writing keys: %v", err)
	}

	// Read them back.
	data, err = ioutil.ReadFile(filepath.Join(dir, "public.upspinkey"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != public {
		t.Fatalf("reading public key: got %q; want %q", data, public)
	}
	data, err = ioutil.ReadFile(filepath.Join(dir, "secret.upspinkey"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != private {
		t.Fatalf("reading secret key: got %q; want %q", data, private)
	}

	// Now check the archive.
	data, err = ioutil.ReadFile(filepath.Join(dir, "secret2.upspinkey"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != archive2Key {
		t.Fatalf("reading archive key: got\n%s\n\twant\n%s", data, archive2Key)
	}
}
