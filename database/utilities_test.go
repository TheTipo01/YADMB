package database

import "testing"

func TestEncodeSegments(t *testing.T) {
	result := EncodeSegments(map[int]bool{0: true, 1: true})
	if result != "0,1" {
		t.Error("Encoding segments failed. Expected 0,1, got", result)
	}
}

func TestDecodeSegments(t *testing.T) {
	result := DecodeSegments("0,1")
	if result == nil || len(result) == 0 {
		t.Error("Decoding segments failed. Expected map, got nil")
	}
}
