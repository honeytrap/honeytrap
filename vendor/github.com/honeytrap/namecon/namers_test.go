package namecon_test

import (
	"strings"
	"testing"

	"github.com/honeytrap/namecon"
)

// TestSimpleNamer validates the use of the provided namer to match the
// giving template rule set provided.
func TestSimpleNamer(t *testing.T) {
	namer := namecon.GenerateNamer(namecon.SimpleNamer{}, "API-%s-%s")

	firstName := namer("Trappa")
	secondName := namer("Honey")

	if firstName == secondName {
		t.Fatalf("Should have successfully generated new unique generation names")
	}
	t.Logf("Should have successfully generated new unique generation names")
}

// TestLimitedNamer validates the use of the provided namer to match the
// giving template rule set provided.
func TestLimitedNamer(t *testing.T) {
	namer := namecon.GenerateNamer(namecon.NewLimitNamer(50, 10), "API-%s-%s")

	newName := namer("TrappaHouseOfSundaySchool")
	if len(newName) > 50 {
		t.Fatalf("Should have successfully new name within 120 characters")
	}
	t.Logf("Should have successfully new name within 120 characters")

	parts := strings.Split(newName, "-")
	if len(parts[1]) > 10 {
		t.Fatalf("Should have successfully generated base name piece under 20 characters")
	}
	t.Logf("Should have successfully generated base name piece under 20 characters")
}
