package model

import "testing"

func TestModelFromSerial(t *testing.T) {
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

func TestUpsertRegisteredDevice(t *testing.T) {
	u := &User{ID: "u1"}
	u.UpsertRegisteredDevice("RM110ABCDEF", "reMarkable 2", "")
	if len(u.RegisteredDevices) != 1 {
		t.Fatalf("want 1 device, got %d", len(u.RegisteredDevices))
	}
	d := u.RegisteredDevices[0]
	if d.Model != "reMarkable 2" {
		t.Fatalf("model = %q", d.Model)
	}
	u.UpsertRegisteredDevice("RM110ABCDEF", "reMarkable 2", "")
	if len(u.RegisteredDevices) != 1 {
		t.Fatalf("upsert should update in place, got %d", len(u.RegisteredDevices))
	}
	u.RemoveRegisteredDevice("RM110ABCDEF")
	if len(u.RegisteredDevices) != 0 {
		t.Fatalf("want 0 after remove, got %d", len(u.RegisteredDevices))
	}
}
