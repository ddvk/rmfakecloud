package integrations

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/apognu/gocal"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
)

func init() {
	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		return loadTimezone(s), nil
	})
}

var icsCacheMap sync.Map

type cachedICS struct {
	mu        sync.Mutex
	data      []byte
	fetchedAt time.Time
}

type icsIntegration struct {
	url      string
	insecure bool
}

func newICS(cfg model.IntegrationConfig) *icsIntegration {
	return &icsIntegration{
		url:      cfg.Address,
		insecure: cfg.Insecure,
	}
}

func (i *icsIntegration) ListEvents(windowStart, windowEnd time.Time) (*messages.CalendarEventsResponse, error) {
	logrus.Infof("[ics] fetching events from %s, window %s to %s", i.url, windowStart.Format(time.RFC3339), windowEnd.Format(time.RFC3339))

	data, err := i.fetch()
	if err != nil {
		return nil, err
	}

	c := gocal.NewParser(bytes.NewReader(data))
	c.Start = &windowStart
	c.End = &windowEnd
	c.Strict.Mode = gocal.StrictModeFailEvent

	if err := c.Parse(); err != nil {
		return nil, fmt.Errorf("parsing ICS: %w", err)
	}

	logrus.Infof("[ics] parsed %d events in window", len(c.Events))

	var events []messages.CalendarEvent
	for _, e := range c.Events {
		evt := convertEvent(e)
		logrus.Infof("[ics] -> %q start=%s end=%s (utc: %s - %s)", evt.Title, evt.StartTime, evt.EndTime, evt.StartTimeUtc, evt.EndTimeUtc)
		events = append(events, evt)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTimeUtc < events[j].StartTimeUtc
	})

	logrus.Infof("[ics] returning %d events", len(events))
	return &messages.CalendarEventsResponse{
		RetrievedAtUtc: time.Now().UTC().Format("2006-01-02T15:04:05.999999999Z"),
		Events:         events,
	}, nil
}

const icsCacheTTL = 5 * time.Minute

func (i *icsIntegration) fetch() ([]byte, error) {
	val, _ := icsCacheMap.LoadOrStore(i.url, &cachedICS{})
	cached := val.(*cachedICS)

	cached.mu.Lock()
	defer cached.mu.Unlock()

	if cached.data != nil && time.Since(cached.fetchedAt) < icsCacheTTL {
		logrus.Debug("[ics] using cached ICS data")
		return cached.data, nil
	}

	logrus.Infof("[ics] fetching ICS from %s", i.url)
	client := &http.Client{Timeout: 30 * time.Second}
	if i.insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(i.url)
	if err != nil {
		return nil, fmt.Errorf("fetching ICS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching ICS: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ICS: %w", err)
	}

	cached.data = data
	cached.fetchedAt = time.Now()

	logrus.Infof("[ics] fetched %d bytes", len(data))
	return data, nil
}

func convertEvent(e gocal.Event) messages.CalendarEvent {
	eventID := e.Uid
	if e.IsRecurring && e.Start != nil {
		eventID = fmt.Sprintf("%s_%s", e.Uid, e.Start.UTC().Format("20060102T150405Z"))
	}

	var lastMod string
	if e.LastModified != nil {
		lastMod = e.LastModified.UTC().Format("2006-01-02T15:04:05Z")
	} else if e.Stamp != nil {
		lastMod = e.Stamp.UTC().Format("2006-01-02T15:04:05Z")
	} else {
		lastMod = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}

	var startUtc, endUtc, startLocal, endLocal string
	if e.Start != nil {
		startUtc = e.Start.UTC().Format("2006-01-02T15:04:05Z")
		startLocal = e.Start.Format("2006-01-02T15:04:05-07:00")
	}
	if e.End != nil {
		endUtc = e.End.UTC().Format("2006-01-02T15:04:05Z")
		endLocal = e.End.Format("2006-01-02T15:04:05-07:00")
	}

	var organizer *messages.CalendarPerson
	if e.Organizer != nil {
		email := strings.TrimPrefix(strings.ToLower(e.Organizer.Value), "mailto:")
		name := e.Organizer.Cn
		if name == "" {
			name = email
		}
		organizer = &messages.CalendarPerson{Name: name, Email: email}
	}

	var attendees []messages.CalendarPerson
	for _, a := range e.Attendees {
		email := strings.TrimPrefix(strings.ToLower(a.Value), "mailto:")
		name := a.Cn
		if name == "" {
			name = email
		}
		attendees = append(attendees, messages.CalendarPerson{Name: name, Email: email})
	}

	responseStatus := "accepted"
	for _, a := range e.Attendees {
		switch strings.ToUpper(a.Status) {
		case "ACCEPTED":
			responseStatus = "accepted"
		case "DECLINED":
			responseStatus = "declined"
		case "TENTATIVE":
			responseStatus = "tentative"
		}
	}

	return messages.CalendarEvent{
		ID:                  eventID,
		Title:               e.Summary,
		StartTimeUtc:        startUtc,
		EndTimeUtc:          endUtc,
		LastModifiedTimeUtc: lastMod,
		StartTime:           startLocal,
		EndTime:             endLocal,
		ResponseStatus:      responseStatus,
		Organizer:           organizer,
		Attendees:           attendees,
	}
}

var windowsTimezones = map[string]string{
	"AUS Central Standard Time":       "Australia/Darwin",
	"AUS Eastern Standard Time":       "Australia/Sydney",
	"Afghanistan Standard Time":       "Asia/Kabul",
	"Alaskan Standard Time":           "America/Anchorage",
	"Arab Standard Time":              "Asia/Riyadh",
	"Arabian Standard Time":           "Asia/Dubai",
	"Arabic Standard Time":            "Asia/Baghdad",
	"Argentina Standard Time":         "America/Argentina/Buenos_Aires",
	"Atlantic Standard Time":          "America/Halifax",
	"Azerbaijan Standard Time":        "Asia/Baku",
	"Azores Standard Time":            "Atlantic/Azores",
	"Canada Central Standard Time":    "America/Regina",
	"Cape Verde Standard Time":        "Atlantic/Cape_Verde",
	"Central America Standard Time":   "America/Guatemala",
	"Central Asia Standard Time":      "Asia/Almaty",
	"Central Brazilian Standard Time": "America/Cuiaba",
	"Central Europe Standard Time":    "Europe/Budapest",
	"Central European Standard Time":  "Europe/Warsaw",
	"Central Pacific Standard Time":   "Pacific/Guadalcanal",
	"Central Standard Time":           "America/Chicago",
	"Central Standard Time (Mexico)":  "America/Mexico_City",
	"China Standard Time":             "Asia/Shanghai",
	"E. Africa Standard Time":         "Africa/Nairobi",
	"E. Australia Standard Time":      "Australia/Brisbane",
	"E. Europe Standard Time":         "Europe/Chisinau",
	"E. South America Standard Time":  "America/Sao_Paulo",
	"Eastern Standard Time":           "America/New_York",
	"Egypt Standard Time":             "Africa/Cairo",
	"FLE Standard Time":               "Europe/Kiev",
	"GMT Standard Time":               "Europe/London",
	"GTB Standard Time":               "Europe/Bucharest",
	"Georgian Standard Time":          "Asia/Tbilisi",
	"Greenland Standard Time":         "America/Godthab",
	"Greenwich Standard Time":         "Atlantic/Reykjavik",
	"Hawaiian Standard Time":          "Pacific/Honolulu",
	"India Standard Time":             "Asia/Calcutta",
	"Iran Standard Time":              "Asia/Tehran",
	"Israel Standard Time":            "Asia/Jerusalem",
	"Jordan Standard Time":            "Asia/Amman",
	"Korea Standard Time":             "Asia/Seoul",
	"Mauritius Standard Time":         "Indian/Mauritius",
	"Middle East Standard Time":       "Asia/Beirut",
	"Mountain Standard Time":          "America/Denver",
	"Mountain Standard Time (Mexico)": "America/Chihuahua",
	"Myanmar Standard Time":           "Asia/Rangoon",
	"N. Central Asia Standard Time":   "Asia/Novosibirsk",
	"Namibia Standard Time":           "Africa/Windhoek",
	"Nepal Standard Time":             "Asia/Katmandu",
	"New Zealand Standard Time":       "Pacific/Auckland",
	"Newfoundland Standard Time":      "America/St_Johns",
	"North Asia East Standard Time":   "Asia/Irkutsk",
	"North Asia Standard Time":        "Asia/Krasnoyarsk",
	"Pacific SA Standard Time":        "America/Santiago",
	"Pacific Standard Time":           "America/Los_Angeles",
	"Pacific Standard Time (Mexico)":  "America/Tijuana",
	"Pakistan Standard Time":          "Asia/Karachi",
	"Romance Standard Time":           "Europe/Paris",
	"Russia Time Zone 3":              "Europe/Samara",
	"Russian Standard Time":           "Europe/Moscow",
	"SA Eastern Standard Time":        "America/Cayenne",
	"SA Pacific Standard Time":        "America/Bogota",
	"SA Western Standard Time":        "America/La_Paz",
	"SE Asia Standard Time":           "Asia/Bangkok",
	"Singapore Standard Time":         "Asia/Singapore",
	"South Africa Standard Time":      "Africa/Johannesburg",
	"Sri Lanka Standard Time":         "Asia/Colombo",
	"Taipei Standard Time":            "Asia/Taipei",
	"Tasmania Standard Time":          "Australia/Hobart",
	"Tokyo Standard Time":             "Asia/Tokyo",
	"Turkey Standard Time":            "Europe/Istanbul",
	"US Eastern Standard Time":        "America/Indianapolis",
	"US Mountain Standard Time":       "America/Phoenix",
	"UTC":                             "UTC",
	"Venezuela Standard Time":         "America/Caracas",
	"W. Australia Standard Time":      "Australia/Perth",
	"W. Central Africa Standard Time": "Africa/Lagos",
	"W. Europe Standard Time":         "Europe/Berlin",
	"West Asia Standard Time":         "Asia/Tashkent",
	"West Pacific Standard Time":      "Pacific/Port_Moresby",
	"Yakutsk Standard Time":           "Asia/Yakutsk",
}

func loadTimezone(tzid string) *time.Location {
	if loc, err := time.LoadLocation(tzid); err == nil {
		return loc
	}

	if iana, ok := windowsTimezones[tzid]; ok {
		if loc, err := time.LoadLocation(iana); err == nil {
			logrus.Debugf("[ics] mapped Windows timezone %q -> %q", tzid, iana)
			return loc
		}
	}

	logrus.Warnf("[ics] unknown timezone %q, falling back to UTC", tzid)
	return time.UTC
}
