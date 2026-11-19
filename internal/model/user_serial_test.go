package model

import "testing"

func TestModelFromSerialMappings(t *testing.T) {
	tests := []struct {
		serial string
		model  string
		ok     bool
	}{
		{serial: "RM02A123456", model: "reMarkable Paper Pro", ok: true},
		{serial: "RM03A999999", model: "reMarkable Paper Pro Move", ok: true},
		{serial: "RM110ABCDEF", model: "reMarkable 2", ok: true},
		{serial: "RM102000001", model: "reMarkable 1", ok: true},
		{serial: "RM12A111111", model: "TBA", ok: true},
		{serial: "UNKNOWN123", model: "", ok: false},
	}
	for _, tt := range tests {
		got, ok := modelFromSerial(tt.serial)
		if ok != tt.ok {
			t.Fatalf("serial %q: expected ok=%v, got %v", tt.serial, tt.ok, ok)
		}
		if got != tt.model {
			t.Fatalf("serial %q: expected model=%q, got %q", tt.serial, tt.model, got)
		}
	}
}
