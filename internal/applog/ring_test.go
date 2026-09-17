package applog

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestRingSnapshotOrder(t *testing.T) {
	r := New(3)
	r.Add(Line{Time: time.Now(), Level: "info", Message: "a"})
	r.Add(Line{Time: time.Now(), Level: "info", Message: "b"})
	r.Add(Line{Time: time.Now(), Level: "info", Message: "c"})
	r.Add(Line{Time: time.Now(), Level: "warn", Message: "d"})
	got := r.Texts(0)
	if len(got) != 3 {
		t.Fatalf("len %d", len(got))
	}
	if !strings.Contains(got[0], "b") || !strings.Contains(got[1], "c") || !strings.Contains(got[2], "d") {
		t.Fatalf("%v", got)
	}
	got2 := r.Texts(2)
	if len(got2) != 2 || !strings.Contains(got2[1], "d") {
		t.Fatalf("%v", got2)
	}
}

func TestHookFire(t *testing.T) {
	r := New(10)
	h := &Hook{Ring: r}
	e := logrus.WithField("user", "x")
	e.Time = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	e.Level = logrus.InfoLevel
	e.Message = "hello"
	if err := h.Fire(e); err != nil {
		t.Fatal(err)
	}
	lines := r.Snapshot(0)
	if len(lines) != 1 || lines[0].Message != "hello user=x" {
		t.Fatalf("%+v", lines)
	}
}
