package integrations

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/apognu/gocal"
)

func parseTestEvents(t *testing.T, data string, start, end time.Time) []gocal.Event {
	t.Helper()
	c := gocal.NewParser(bytes.NewReader([]byte(data)))
	c.Start = &start
	c.End = &end
	c.Strict.Mode = gocal.StrictModeFailEvent
	if err := c.Parse(); err != nil {
		t.Fatalf("parsing test calendar: %v", err)
	}
	return c.Events
}

func TestSingleEvent(t *testing.T) {
	start := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:single-event-1
SUMMARY:Team Meeting
DTSTART:20260325T160000Z
DTEND:20260325T170000Z
DTSTAMP:20260325T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := convertEvent(events[0])
	if evt.ID != "single-event-1" {
		t.Errorf("expected ID single-event-1, got %s", evt.ID)
	}
	if evt.Title != "Team Meeting" {
		t.Errorf("expected title Team Meeting, got %s", evt.Title)
	}
	if evt.StartTimeUtc != "2026-03-25T16:00:00Z" {
		t.Errorf("unexpected StartTimeUtc: %s", evt.StartTimeUtc)
	}
	if evt.EndTimeUtc != "2026-03-25T17:00:00Z" {
		t.Errorf("unexpected EndTimeUtc: %s", evt.EndTimeUtc)
	}
}

func TestEventOutsideWindow(t *testing.T) {
	start := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:outside-1
SUMMARY:Old Event
DTSTART:20260301T160000Z
DTEND:20260301T170000Z
DTSTAMP:20260301T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 0 {
		t.Fatalf("expected 0 events for out-of-window, got %d", len(events))
	}
}

func TestWeeklyRecurrence(t *testing.T) {
	start := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:weekly-1
SUMMARY:Weekly Standup
DTSTART:20260323T090000Z
DTEND:20260323T093000Z
RRULE:FREQ=WEEKLY;BYDAY=MO
DTSTAMP:20260320T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 2 {
		t.Fatalf("expected 2 occurrences (Mar 23, Mar 30), got %d", len(events))
	}

	evt0 := convertEvent(events[0])
	evt1 := convertEvent(events[1])

	if !strings.HasPrefix(evt0.ID, "weekly-1_") {
		t.Errorf("recurring event ID should have timestamp suffix, got %s", evt0.ID)
	}
	if evt0.StartTimeUtc != "2026-03-23T09:00:00Z" {
		t.Errorf("first occurrence: %s", evt0.StartTimeUtc)
	}
	if evt1.StartTimeUtc != "2026-03-30T09:00:00Z" {
		t.Errorf("second occurrence: %s", evt1.StartTimeUtc)
	}
	if evt0.EndTimeUtc != "2026-03-23T09:30:00Z" {
		t.Errorf("duration not preserved: %s", evt0.EndTimeUtc)
	}
}

func TestExdateExclusion(t *testing.T) {
	start := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:exdate-1
SUMMARY:Daily Sync
DTSTART:20260323T140000Z
DTEND:20260323T143000Z
RRULE:FREQ=DAILY;COUNT=5
EXDATE:20260325T140000Z
DTSTAMP:20260320T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 4 {
		t.Fatalf("expected 4 occurrences (5 minus 1 EXDATE), got %d", len(events))
	}

	for _, e := range events {
		evt := convertEvent(e)
		if evt.StartTimeUtc == "2026-03-25T14:00:00Z" {
			t.Error("EXDATE March 25 should have been excluded")
		}
	}
}

func TestAllDayEvent(t *testing.T) {
	start := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:allday-1
SUMMARY:Company Holiday
DTSTART;VALUE=DATE:20260325
DTEND;VALUE=DATE:20260326
DTSTAMP:20260320T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 1 {
		t.Fatalf("expected 1 all-day event, got %d", len(events))
	}

	evt := convertEvent(events[0])
	if evt.StartTimeUtc != "2026-03-25T00:00:00Z" {
		t.Errorf("all-day start: %s", evt.StartTimeUtc)
	}
}

func TestTimezone(t *testing.T) {
	start := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:tz-1
SUMMARY:Denver Meeting
DTSTART;TZID=America/Denver:20260325T103000
DTEND;TZID=America/Denver:20260325T110000
DTSTAMP:20260320T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := convertEvent(events[0])
	if evt.StartTimeUtc != "2026-03-25T16:30:00Z" {
		t.Errorf("expected UTC conversion 16:30, got %s", evt.StartTimeUtc)
	}
	if evt.StartTime != "2026-03-25T10:30:00-06:00" {
		t.Errorf("expected local time with offset, got %s", evt.StartTime)
	}
}

func TestOrganizerAndAttendees(t *testing.T) {
	start := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:attendee-test
SUMMARY:Meeting
DTSTART:20260325T160000Z
DTEND:20260325T170000Z
DTSTAMP:20260325T100000Z
ORGANIZER;CN=Alice Smith:mailto:alice@example.com
ATTENDEE;CN=Bob Jones;PARTSTAT=ACCEPTED:mailto:bob@example.com
ATTENDEE;CN=Carol White;PARTSTAT=TENTATIVE:mailto:carol@example.com
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := convertEvent(events[0])
	if evt.Organizer == nil {
		t.Fatal("expected organizer")
	}
	if evt.Organizer.Name != "Alice Smith" {
		t.Errorf("organizer name: %s", evt.Organizer.Name)
	}
	if evt.Organizer.Email != "alice@example.com" {
		t.Errorf("organizer email: %s", evt.Organizer.Email)
	}
	if len(evt.Attendees) != 2 {
		t.Fatalf("expected 2 attendees, got %d", len(evt.Attendees))
	}
	if evt.Attendees[0].Name != "Bob Jones" {
		t.Errorf("attendee 0 name: %s", evt.Attendees[0].Name)
	}
}

func TestMalformedEventSkipped(t *testing.T) {
	start := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	events := parseTestEvents(t, `BEGIN:VCALENDAR
BEGIN:VEVENT
UID:good-event
SUMMARY:Good Event
DTSTART:20260325T160000Z
DTEND:20260325T170000Z
DTSTAMP:20260325T100000Z
END:VEVENT
BEGIN:VEVENT
UID:bad-event
SUMMARY:Bad Event
DTSTART;TZID=America/Denver:20260325T103000
DTEND;TZID=America/Denver:20260325T110000
DTSTAMP:20260320T100000Z
X-APPLE-STRUCTURED-LOCATION;VALUE=URI;X-TITLE=123 Main St.
Bad Unfolded Line
Another Bad Line:geo:35.0,-95.0
END:VEVENT
BEGIN:VEVENT
UID:good-event-2
SUMMARY:Also Good
DTSTART:20260325T180000Z
DTEND:20260325T190000Z
DTSTAMP:20260325T100000Z
END:VEVENT
END:VCALENDAR`, start, end)

	if len(events) < 1 {
		t.Fatalf("expected at least 1 good event, got %d", len(events))
	}

	var titles []string
	for _, e := range events {
		titles = append(titles, e.Summary)
	}
	found := false
	for _, title := range titles {
		if title == "Good Event" || title == "Also Good" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected good events to survive, got titles: %v", titles)
	}
}
