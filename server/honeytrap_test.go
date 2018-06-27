package server

import (
	"testing"
)

func TestBigPortToAddr(t *testing.T) {
	addr, proto, port, err := ToAddr("tcp/60000")
	if err != nil {
		t.Fatal(err)
	}

	if addr.String() != ":60000" {
		t.Errorf("Expected :60000 but got %s", addr)
	}
	if proto != "tcp" {
		t.Errorf("Expected tcp but got %s", proto)
	}
	if port != 60000 {
		t.Errorf("Expected 60000 but got %d", port)
	}
}

func TestIncorrectSeparatorToAddr(t *testing.T) {
	_, _, _, err := ToAddr("tcp:8080")
	if err == nil {
		t.Errorf("No error thrown with incorrect separator")
	}
}


func TestUnknownProtoToAddr(t *testing.T) {
	_, _, _, err := ToAddr("tdp:8080")
	if err == nil {
		t.Errorf("No error thrown with incorrect protocol")
	}
}
