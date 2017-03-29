package controller

import (
	"encoding/json"
	"testing"
)

func BenchmarkJSONString(t *testing.B) {
	var s = "the need for mead"
	for i := 0; i < t.N; i++ {
		jsonString(s)
	}
}

func BenchmarkJSONMarshal(t *testing.B) {
	var s = "the need for mead"
	for i := 0; i < t.N; i++ {
		json.Marshal(s)
	}
}
