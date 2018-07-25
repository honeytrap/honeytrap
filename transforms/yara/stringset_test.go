package yara

import (
	"testing"
)

func TestHas(t *testing.T) {
	a := make(stringSet)
	if a.Has("foo") {
		t.Error("Has a non-existing key")
	}
	a.Add("foo")
	if !a.Has("foo") {
		t.Error("Doesn't have a set key")
	}
	a.Remove("foo")
	if a.Has("foo") {
		t.Error("Has a removed key")
	}
}

func TestRemove(t *testing.T) {
	a := make(stringSet)
	a.Add("foo")
	a.Remove("foo")
	if a.Has("foo") {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	a := make(stringSet)
	a.Add("foo")
	if !a.Has("foo") {
		t.Fail()
	}
}

func TestMerge(t *testing.T) {
	a := make(stringSet)
	b := make(stringSet)
	a.Add("foo")
	b.Add("bar")
	result := a.Merge(b)
	if !a.Has("foo") || !a.Has("bar") {
		t.Fail()
	}
	if !result.Has("foo") || !result.Has("bar") {
		t.Fail()
	}
}
