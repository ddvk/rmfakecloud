package applog

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const defaultCapacity = 500

// Line is one captured log entry for the admin UI.
type Line struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

func (l Line) Text() string {
	ts := l.Time.Local().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s %-7s %s", ts, strings.ToUpper(l.Level), l.Message)
}

// Ring is a fixed-size buffer of recent log lines.
type Ring struct {
	mu   sync.RWMutex
	buf  []Line
	cap  int
	next int
	full bool
}

var (
	defaultRing     *Ring
	defaultRingOnce sync.Once
)

// Default returns the process-wide ring used by the admin panel.
func Default() *Ring {
	defaultRingOnce.Do(func() {
		defaultRing = New(defaultCapacity)
	})
	return defaultRing
}

// New creates a ring that keeps the last capacity lines.
func New(capacity int) *Ring {
	if capacity < 1 {
		capacity = defaultCapacity
	}
	return &Ring{
		buf: make([]Line, capacity),
		cap: capacity,
	}
}

// Add appends a line.
func (r *Ring) Add(line Line) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.next] = line
	r.next = (r.next + 1) % r.cap
	if r.next == 0 {
		r.full = true
	}
}

// Snapshot returns the newest lines (oldest first), at most limit (0 = all).
func (r *Ring) Snapshot(limit int) []Line {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := r.next
	if r.full {
		n = r.cap
	}
	out := make([]Line, 0, n)
	if !r.full {
		out = append(out, r.buf[:r.next]...)
	} else {
		out = append(out, r.buf[r.next:]...)
		out = append(out, r.buf[:r.next]...)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// Texts returns Snapshot lines formatted as plain text.
func (r *Ring) Texts(limit int) []string {
	lines := r.Snapshot(limit)
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.Text()
	}
	return out
}

// Hook implements logrus.Hook and feeds Default (or a custom ring).
type Hook struct {
	Ring *Ring
}

// Levels captures all levels.
func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire stores the entry.
func (h *Hook) Fire(e *logrus.Entry) error {
	r := h.Ring
	if r == nil {
		r = Default()
	}
	msg := e.Message
	if len(e.Data) > 0 {
		parts := make([]string, 0, len(e.Data))
		for k, v := range e.Data {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		msg = msg + " " + strings.Join(parts, " ")
	}
	r.Add(Line{
		Time:    e.Time,
		Level:   e.Level.String(),
		Message: msg,
	})
	return nil
}

// InstallHook registers the default ring on the standard logger.
func InstallHook() {
	logrus.AddHook(&Hook{Ring: Default()})
}
