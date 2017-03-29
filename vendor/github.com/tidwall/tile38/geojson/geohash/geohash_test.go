package geohash

import (
	"fmt"
	"testing"
)

func fixed(f float64, d int) string {
	return fmt.Sprintf(fmt.Sprintf("%%0.%df", d), f)
}

func TestABC(t *testing.T) {
	lat, lon := 33.52345123, -115.512345123
	hash, err := Encode(lat, lon, 32)
	if err != nil {
		t.Fatal(err)
	}
	lat2, lon2, err := Decode(hash)
	if err != nil {
		t.Fatal(err)
	}
	if fixed(lat, 10) != fixed(lat2, 10) || fixed(lon, 10) != fixed(lon2, 10) {
		t.Fatalf("bad geohash %v,%v %v,%v", lat, lon, lat2, lon2)
	}
}

// TestEqualsWebserviceHash checks whether an encoded geohash is equal to a
// geohash encoded by geohash.org for identical lat/lon values.
func TestEqualsWebserviceHash(t *testing.T) {
	lat, lon := 27.173117, 78.042122
	hash, err := Encode(lat, lon, 12)
	if err != nil {
		t.Fatal(err)
	}

	hash2 := "tsz6xfswchpu"
	if hash != hash2 {
		t.Errorf("geohash should be equal %v, %v", hash, hash2)
	}
}

func TestNearbyHasCommonPrefix(t *testing.T) {
	lat, lon := 27.174583139355413, 78.04258346557617
	hash, err := Encode(lat, lon, 32)
	if err != nil {
		t.Fatal(err)
	}

	lat2, lon2 := 27.174559277910305, 78.04163932800293
	hash2, err := Encode(lat2, lon2, 32)
	if err != nil {
		t.Fatal(err)
	}

	// common prefix should be at least of length 7
	pref := hash[:7]
	pref2 := hash2[:7]
	if pref != pref2 {
		t.Errorf("prefix should be equal %v, %v", pref, pref2)
	}
}
