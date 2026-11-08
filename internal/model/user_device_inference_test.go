package model

import "testing"

func TestModelFromSerial(t *testing.T) {
	tests := []struct {
		serial string
		want   string
	}{
		{"RM02A12345", "reMarkable Paper Pro"},
		{"RM03A-9999", "reMarkable Paper Pro Move"},
		{"RM110ABCDE", "reMarkable 2"},
		{"RM102ABCDE", "reMarkable 1"},
		{"RM12A00001", "TBA"},
	}
	for _, tt := range tests {
		got, ok := modelFromSerial(tt.serial)
		if !ok {
			t.Fatalf("expected mapping for %q", tt.serial)
		}
		if got != tt.want {
			t.Fatalf("serial %q => %q, want %q", tt.serial, got, tt.want)
		}
	}
}
