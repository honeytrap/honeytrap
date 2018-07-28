package yara

import (
	"testing"
)

func TestIdentity(t *testing.T) {
	testCases := []string{
		"foobar",
		"http.url",
		"http.header.user-agent",
	}
	for _, testCase := range testCases {
		if denormalize(normalize(testCase)) != testCase {
			t.Errorf("Identity doesn't apply for %s", testCase)
		}
	}
}