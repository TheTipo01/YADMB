package sponsorblock

import "testing"

func TestHash(t *testing.T) {
	result := hash("ifI_fwg55k8")
	if result != "5f6b" {
		t.Error("Hashing failed. Expected 5f6b, got", result)
	}
}

func TestGetSegments(t *testing.T) {
	result := GetSegments("kJQP7kiw5Fk")
	if result == nil || len(result) == 0 {
		t.Error("Getting segments failed. Expected map, got nil")
	}
}
