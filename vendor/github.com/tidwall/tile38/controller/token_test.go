package controller

import (
	"strings"
	"testing"
)

func TestLowerCompare(t *testing.T) {
	if !lc("hello", "hello") {
		t.Fatal("failed")
	}
	if !lc("Hello", "hello") {
		t.Fatal("failed")
	}
	if !lc("HeLLo World", "hello world") {
		t.Fatal("failed")
	}
	if !lc("", "") {
		t.Fatal("failed")
	}
	if lc("hello", "") {
		t.Fatal("failed")
	}
	if lc("", "hello") {
		t.Fatal("failed")
	}
	if lc("HeLLo World", "Hello world") {
		t.Fatal("failed")
	}
}

// func testParseFloat(t testing.TB, s string, f float64, invalid bool) {
// 	n, err := parseFloat(s)
// 	if err != nil {
// 		if invalid {
// 			return
// 		}
// 		t.Fatal(err)
// 	}
// 	if invalid {
// 		t.Fatalf("expecting an error for %s", s)
// 	}
// 	if n != f {
// 		t.Fatalf("for '%s', expect %f, got %f", s, f, n)
// 	}
// }

// func TestParseFloat(t *testing.T) {
// 	testParseFloat(t, "100", 100, false)
// 	testParseFloat(t, "0", 0, false)
// 	testParseFloat(t, "-1", -1, false)
// 	testParseFloat(t, "-0", -0, false)

// 	testParseFloat(t, "-100", -100, false)
// 	testParseFloat(t, "-0", -0, false)
// 	testParseFloat(t, "+1", 1, false)
// 	testParseFloat(t, "+0", 0, false)

// 	testParseFloat(t, "33.102938", 33.102938, false)
// 	testParseFloat(t, "-115.123123", -115.123123, false)

// 	testParseFloat(t, ".1", 0.1, false)
// 	testParseFloat(t, "0.1", 0.1, false)

// 	testParseFloat(t, "00.1", 0.1, false)
// 	testParseFloat(t, "01.1", 1.1, false)
// 	testParseFloat(t, "01", 1, false)
// 	testParseFloat(t, "-00.1", -0.1, false)
// 	testParseFloat(t, "+00.1", 0.1, false)
// 	testParseFloat(t, "", 0.1, true)
// 	testParseFloat(t, " 0", 0.1, true)
// 	testParseFloat(t, "0 ", 0.1, true)

// }

func BenchmarkLowerCompare(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if !lc("HeLLo World", "hello world") {
			t.Fatal("failed")
		}
	}
}

func BenchmarkStringsLowerCompare(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if strings.ToLower("HeLLo World") != "hello world" {
			t.Fatal("failed")
		}

	}
}

// func BenchmarkParseFloat(t *testing.B) {
// 	s := []string{"33.10293", "-115.1203102"}
// 	for i := 0; i < t.N; i++ {
// 		_, err := parseFloat(s[i%2])
// 		if err != nil {
// 			t.Fatal("failed")
// 		}
// 	}
// }

// func BenchmarkStrconvParseFloat(t *testing.B) {
// 	s := []string{"33.10293", "-115.1203102"}
// 	for i := 0; i < t.N; i++ {
// 		_, err := strconv.ParseFloat(s[i%2], 64)
// 		if err != nil {
// 			t.Fatal("failed")
// 		}
// 	}
// }
