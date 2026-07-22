package database

import "testing"

func TestEncodeSegments(t *testing.T) {
	result := EncodeSegments(map[int]struct{}{0: {}, 1: {}})
	if result != "0,1" {
		t.Error("Encoding segments failed. Expected 0,1, got", result)
	}
}

func TestDecodeSegments(t *testing.T) {
	result := DecodeSegments("0,1")
	if len(result) == 0 {
		t.Error("Decoding segments failed. Expected map, got nil")
	}
}
