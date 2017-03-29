package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	f := &bytes.Buffer{}
	Default = New(f, &Config{})
	Printf("hello %v", "everyone")
	if !strings.HasSuffix(f.String(), "hello everyone\n") {
		t.Fatal("fail")
	}
}
