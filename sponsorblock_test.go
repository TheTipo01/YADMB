package main

import "testing"

func TestHash(t *testing.T) {
	result := hash("ifI_fwg55k8")
	if result != "5f6b" {
		t.Error("Hashing failed. Expected 5f6b, got", result)
	}
}

func TestGetSegments(t *testing.T) {
	result := getSegments("3IJ0Lk2-w_I")
	if result == nil || len(result) == 0 {
		t.Error("Getting segments failed. Expected map, got nil")
	}
}

func TestEncodeSegments(t *testing.T) {
	result := encodeSegments(map[int]bool{0: true, 1: true})
	if result != "0,1" {
		t.Error("Encoding segments failed. Expected 0,1, got", result)
	}
}

func TestDecodeSegments(t *testing.T) {
	result := decodeSegments("0,1")
	if result == nil || len(result) == 0 {
		t.Error("Decoding segments failed. Expected map, got nil")
	}
}
