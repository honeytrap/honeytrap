package glob

import (
	"math/rand"
	"testing"
	"time"
)

func test(t *testing.T, pattern string, desc bool, limitsExpect []string, isGlobExpect bool) {
	g := Parse(pattern, desc)
	if g.IsGlob != isGlobExpect {
		t.Fatalf("pattern[%v] desc[%v] (isGlob=%v, expected=%v)", pattern, desc, g.IsGlob, isGlobExpect)
	}
	if g.Limits[0] != limitsExpect[0] || g.Limits[1] != limitsExpect[1] {
		t.Fatalf("pattern[%v] desc[%v] (limits=%v, expected=%v)", pattern, desc, g.Limits, limitsExpect)
	}
	if g.Pattern != pattern {
		t.Fatalf("pattern[%v] desc[%v] (pattern=%v, expected=%v)", pattern, desc, g.Pattern, pattern)
	}
	if g.Desc != desc {
		t.Fatalf("pattern[%v] desc[%v] (desc=%v, expected=%v)", pattern, desc, g.Desc, desc)
	}
}

func TestGlob(t *testing.T) {
	test(t, "*", false, []string{"", ""}, true)
	test(t, "", false, []string{"", ""}, false)
	test(t, "hello*", false, []string{"hello", "hellp"}, true)
	test(t, "hello", false, []string{"hello", "hellp"}, false)
	test(t, "\xff*", false, []string{"\xff", "\xff\x00"}, true)
	test(t, "\x00*", false, []string{"\x00", "\x01"}, true)
	test(t, "\xff", false, []string{"\xff", "\xff\x00"}, false)

	test(t, "*", true, []string{"", ""}, true)
	test(t, "", true, []string{"", ""}, false)
	test(t, "hello*", true, []string{"hellp", "helln"}, true)
	test(t, "hello", true, []string{"hellp", "helln"}, false)
	test(t, "a\xff*", true, []string{"a\xff\x00", "a\xfe"}, true)
	test(t, "\x00*", true, []string{"\x01", ""}, true)
	test(t, "\x01*", true, []string{"\x02", "\x00"}, true)
	test(t, "b\x00*", true, []string{"b\x01", "a\xff"}, true)
	test(t, "\x00\x00*", true, []string{"\x00\x01", ""}, true)
	test(t, "\x00\x01\x00*", true, []string{"\x00\x01\x01", "\x00\x00\xff"}, true)
}

func testMatch(s, pattern string) bool {
	ok, _ := Match(pattern, s)
	return ok
}

func TestMatch(t *testing.T) {
	if !testMatch("hello world", "hello world") {
		t.Fatal("fail")
	}
	if testMatch("hello world", "jello world") {
		t.Fatal("fail")
	}
	if !testMatch("hello world", "hello*") {
		t.Fatal("fail")
	}
	if testMatch("hello world", "jello*") {
		t.Fatal("fail")
	}
	if !testMatch("hello world", "hello?world") {
		t.Fatal("fail")
	}
	if testMatch("hello world", "jello?world") {
		t.Fatal("fail")
	}
	if !testMatch("hello world", "he*o?world") {
		t.Fatal("fail")
	}
	if !testMatch("hello world", "he*o?wor*") {
		t.Fatal("fail")
	}
	if !testMatch("hello world", "he*o?*r*") {
		t.Fatal("fail")
	}
	if !testMatch("的情况下解析一个", "*") {
		t.Fatal("fail")
	}
	if !testMatch("的情况下解析一个", "*况下*") {
		t.Fatal("fail")
	}
	if !testMatch("的情况下解析一个", "*况?*") {
		t.Fatal("fail")
	}
	if !testMatch("的情况下解析一个", "的情况?解析一个") {
		t.Fatal("fail")
	}
}

// TestWildcardMatch - Tests validate the logic of wild card matching.
// `WildcardMatch` supports '*' and '?' wildcards.
// Sample usage: In resource matching for folder policy validation.
func TestWildcardMatch(t *testing.T) {
	testCases := []struct {
		pattern string
		text    string
		matched bool
	}{
		// Test case - 1.
		// Test case with pattern containing key name with a prefix. Should accept the same text without a "*".
		{
			pattern: "my-folder/oo*",
			text:    "my-folder/oo",
			matched: true,
		},
		// Test case - 2.
		// Test case with "*" at the end of the pattern.
		{
			pattern: "my-folder/In*",
			text:    "my-folder/India/Karnataka/",
			matched: true,
		},
		// Test case - 3.
		// Test case with prefixes shuffled.
		// This should fail.
		{
			pattern: "my-folder/In*",
			text:    "my-folder/Karnataka/India/",
			matched: false,
		},
		// Test case - 4.
		// Test case with text expanded to the wildcards in the pattern.
		{
			pattern: "my-folder/In*/Ka*/Ban",
			text:    "my-folder/India/Karnataka/Ban",
			matched: true,
		},
		// Test case - 5.
		// Test case with the  keyname part is repeated as prefix several times.
		// This is valid.
		{
			pattern: "my-folder/In*/Ka*/Ban",
			text:    "my-folder/India/Karnataka/Ban/Ban/Ban/Ban/Ban",
			matched: true,
		},
		// Test case - 6.
		// Test case to validate that `*` can be expanded into multiple prefixes.
		{
			pattern: "my-folder/In*/Ka*/Ban",
			text:    "my-folder/India/Karnataka/Area1/Area2/Area3/Ban",
			matched: true,
		},
		// Test case - 7.
		// Test case to validate that `*` can be expanded into multiple prefixes.
		{
			pattern: "my-folder/In*/Ka*/Ban",
			text:    "my-folder/India/State1/State2/Karnataka/Area1/Area2/Area3/Ban",
			matched: true,
		},
		// Test case - 8.
		// Test case where the keyname part of the pattern is expanded in the text.
		{
			pattern: "my-folder/In*/Ka*/Ban",
			text:    "my-folder/India/Karnataka/Bangalore",
			matched: false,
		},
		// Test case - 9.
		// Test case with prefixes and wildcard expanded for all "*".
		{
			pattern: "my-folder/In*/Ka*/Ban*",
			text:    "my-folder/India/Karnataka/Bangalore",
			matched: true,
		},
		// Test case - 10.
		// Test case with keyname part being a wildcard in the pattern.
		{pattern: "my-folder/*",
			text:    "my-folder/India",
			matched: true,
		},
		// Test case - 11.
		{
			pattern: "my-folder/oo*",
			text:    "my-folder/odo",
			matched: false,
		},

		// Test case with pattern containing wildcard '?'.
		// Test case - 12.
		// "my-folder?/" matches "my-folder1/", "my-folder2/", "my-folder3" etc...
		// doesn't match "myfolder/".
		{
			pattern: "my-folder?/abc*",
			text:    "myfolder/abc",
			matched: false,
		},
		// Test case - 13.
		{
			pattern: "my-folder?/abc*",
			text:    "my-folder1/abc",
			matched: true,
		},
		// Test case - 14.
		{
			pattern: "my-?-folder/abc*",
			text:    "my--folder/abc",
			matched: false,
		},
		// Test case - 15.
		{
			pattern: "my-?-folder/abc*",
			text:    "my-1-folder/abc",
			matched: true,
		},
		// Test case - 16.
		{
			pattern: "my-?-folder/abc*",
			text:    "my-k-folder/abc",
			matched: true,
		},
		// Test case - 17.
		{
			pattern: "my??folder/abc*",
			text:    "myfolder/abc",
			matched: false,
		},
		// Test case - 18.
		{
			pattern: "my??folder/abc*",
			text:    "my4afolder/abc",
			matched: true,
		},
		// Test case - 19.
		{
			pattern: "my-folder?abc*",
			text:    "my-folder/abc",
			matched: true,
		},
		// Test case 20-21.
		// '?' matches '/' too. (works with s3).
		// This is because the namespace is considered flat.
		// "abc?efg" matches both "abcdefg" and "abc/efg".
		{
			pattern: "my-folder/abc?efg",
			text:    "my-folder/abcdefg",
			matched: true,
		},
		{
			pattern: "my-folder/abc?efg",
			text:    "my-folder/abc/efg",
			matched: true,
		},
		// Test case - 22.
		{
			pattern: "my-folder/abc????",
			text:    "my-folder/abc",
			matched: false,
		},
		// Test case - 23.
		{
			pattern: "my-folder/abc????",
			text:    "my-folder/abcde",
			matched: false,
		},
		// Test case - 24.
		{
			pattern: "my-folder/abc????",
			text:    "my-folder/abcdefg",
			matched: true,
		},
		// Test case 25-26.
		// test case with no '*'.
		{
			pattern: "my-folder/abc?",
			text:    "my-folder/abc",
			matched: false,
		},
		{
			pattern: "my-folder/abc?",
			text:    "my-folder/abcd",
			matched: true,
		},
		{
			pattern: "my-folder/abc?",
			text:    "my-folder/abcde",
			matched: false,
		},
		// Test case 27.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnop",
			matched: false,
		},
		// Test case 28.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnopqrst/mnopqr",
			matched: true,
		},
		// Test case 29.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnopqrst/mnopqrs",
			matched: true,
		},
		// Test case 30.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnop",
			matched: false,
		},
		// Test case 31.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnopq",
			matched: true,
		},
		// Test case 32.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnopqr",
			matched: true,
		},
		// Test case 33.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopqand",
			matched: true,
		},
		// Test case 34.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopand",
			matched: false,
		},
		// Test case 35.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopqand",
			matched: true,
		},
		// Test case 36.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mn",
			matched: false,
		},
		// Test case 37.
		{
			pattern: "my-folder/mnop*?",
			text:    "my-folder/mnopqrst/mnopqrs",
			matched: true,
		},
		// Test case 38.
		{
			pattern: "my-folder/mnop*??",
			text:    "my-folder/mnopqrst",
			matched: true,
		},
		// Test case 39.
		{
			pattern: "my-folder/mnop*qrst",
			text:    "my-folder/mnopabcdegqrst",
			matched: true,
		},
		// Test case 40.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopqand",
			matched: true,
		},
		// Test case 41.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopand",
			matched: false,
		},
		// Test case 42.
		{
			pattern: "my-folder/mnop*?and?",
			text:    "my-folder/mnopqanda",
			matched: true,
		},
		// Test case 43.
		{
			pattern: "my-folder/mnop*?and",
			text:    "my-folder/mnopqanda",
			matched: false,
		},
		// Test case 44.

		{
			pattern: "my-?-folder/abc*",
			text:    "my-folder/mnopqanda",
			matched: false,
		},
	}
	// Iterating over the test cases, call the function under test and asert the output.
	for i, testCase := range testCases {
		actualResult := testMatch(testCase.text, testCase.pattern)
		if testCase.matched != actualResult {
			t.Errorf("Test %d: Expected the result to be `%v`, but instead found it to be `%v`", i+1, testCase.matched, actualResult)
		}
	}
}
func TestRandomInput(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	b1 := make([]byte, 100)
	b2 := make([]byte, 100)
	for i := 0; i < 1000000; i++ {
		if _, err := rand.Read(b1); err != nil {
			t.Fatal(err)
		}
		if _, err := rand.Read(b2); err != nil {
			t.Fatal(err)
		}
		testMatch(string(b1), string(b2))
	}
}
func BenchmarkAscii(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if !testMatch("hello", "hello") {
			t.Fatal("fail")
		}
	}
}

func BenchmarkUnicode(t *testing.B) {
	for i := 0; i < t.N; i++ {
		if !testMatch("h情llo", "h情llo") {
			t.Fatal("fail")
		}
	}
}
