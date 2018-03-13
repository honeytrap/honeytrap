package banner

import (
	"strings"
	"testing"
	"time"
)

func TestBannerNoDateTime(t *testing.T) {
	vals := []struct {
		host, text, want string
	}{
		{"abc.com", "Welcome", "abc.com Welcome "},
		{"abc.com", "", "abc.com "},
		{"", "Welcome", "Welcome "},
		{"", "", ""},
	}

	for _, c := range vals {
		tmpl, err := New(c.host, c.text, false)
		if err != nil {
			t.Log(err)
		}

		got := tmpl.Banner()
		if got != c.want {
			t.Errorf("Banner want [%q] got [%q]", c.want, got)
		}
	}
}

func TestBannerWithDateTime(t *testing.T) {
	var want strings.Builder
	want.Grow(64)

	vals := []struct {
		host, text, want string
	}{
		{"abc.com", "Welcome", "abc.com Welcome "},
		{"abc.com", "", "abc.com "},
		{"", "Welcome", "Welcome "},
		{"", "", ""},
	}

	for _, c := range vals {
		// This should run within 1 second, otherwise test fails due to wrong time

		timestring := time.Now().Format(time.RFC3339)
		want.WriteString(c.want)
		want.WriteString(timestring)

		tmpl, err := New(c.host, c.text, true)
		if err != nil {
			t.Log(err)
		}

		got := tmpl.Banner()
		if w := want.String(); w != got {
			t.Errorf("Banner want [%q] got [%q]", w, got)
		}

		want.Reset()
	}
}
