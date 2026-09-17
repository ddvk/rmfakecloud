package model

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	yearRegex       *regexp.Regexp
	serialLikeRegex *regexp.Regexp
)

func init() {
	yearRegex = regexp.MustCompile(`\b(19|20)\d{2}\b`)
	serialLikeRegex = regexp.MustCompile(`\bRM[0-9A-Z]{3,}\b`)
}

// RegisteredDevice is a tablet client that obtained a device token.
type RegisteredDevice struct {
	DeviceID     string    `yaml:"deviceid,omitempty"`
	DeviceDesc   string    `yaml:"devicedesc,omitempty"`
	DeviceLink   string    `yaml:"devicelink,omitempty"`
	Make         string    `yaml:"make,omitempty"`
	Model        string    `yaml:"model,omitempty"`
	Year         string    `yaml:"year,omitempty"`
	RegisteredAt time.Time `yaml:"registeredat,omitempty"`
	LastSeen     time.Time `yaml:"lastseen,omitempty"`
}

// UpsertRegisteredDevice records or updates a paired device for this user.
func (u *User) UpsertRegisteredDevice(deviceID, desc, link string) {
	if u == nil || strings.TrimSpace(deviceID) == "" {
		return
	}
	makeName, model, year := inferDeviceInfo(deviceID, desc, link)
	now := time.Now()
	for i := range u.RegisteredDevices {
		if u.RegisteredDevices[i].DeviceID == deviceID {
			u.RegisteredDevices[i].DeviceDesc = desc
			if link != "" {
				u.RegisteredDevices[i].DeviceLink = link
			}
			if makeName != "" {
				u.RegisteredDevices[i].Make = makeName
			}
			if model != "" {
				u.RegisteredDevices[i].Model = model
			}
			if year != "" {
				u.RegisteredDevices[i].Year = year
			}
			u.RegisteredDevices[i].LastSeen = now
			u.UpdatedAt = now
			return
		}
	}
	u.RegisteredDevices = append(u.RegisteredDevices, RegisteredDevice{
		DeviceID:     deviceID,
		DeviceDesc:   desc,
		DeviceLink:   link,
		Make:         makeName,
		Model:        model,
		Year:         year,
		RegisteredAt: now,
		LastSeen:     now,
	})
	u.UpdatedAt = now
}

// GetRegisteredDevice returns a stored device entry if present.
func (u *User) GetRegisteredDevice(deviceID string) (RegisteredDevice, bool) {
	if u == nil {
		return RegisteredDevice{}, false
	}
	for _, d := range u.RegisteredDevices {
		if d.DeviceID == deviceID {
			return d, true
		}
	}
	return RegisteredDevice{}, false
}

// RemoveRegisteredDevice drops a device from the registry (e.g. tablet logout).
func (u *User) RemoveRegisteredDevice(deviceID string) {
	if u == nil || deviceID == "" {
		return
	}
	j := 0
	for _, d := range u.RegisteredDevices {
		if d.DeviceID != deviceID {
			u.RegisteredDevices[j] = d
			j++
		}
	}
	u.RegisteredDevices = u.RegisteredDevices[:j]
	u.UpdatedAt = time.Now()
}

func inferDeviceInfo(deviceID, desc, link string) (string, string, string) {
	makeName := ""
	model := ""
	year := ""

	ls := strings.ToLower(desc + " " + link)
	if strings.Contains(ls, "remarkable") || strings.Contains(ls, "re-markable") || strings.Contains(ls, "rm2") || strings.Contains(ls, "rm1") {
		makeName = "reMarkable"
	}
	if strings.Contains(ls, "paper pro") || strings.Contains(ls, "paperpro") {
		model = "Paper Pro"
	} else if strings.Contains(ls, "remarkable 2") || strings.Contains(ls, "rm2") {
		model = "2"
	} else if strings.Contains(ls, "remarkable 1") || strings.Contains(ls, "rm1") {
		model = "1"
	}

	if u, err := url.Parse(link); err == nil {
		q := u.Query()
		if v := strings.TrimSpace(q.Get("make")); v != "" {
			makeName = v
		}
		if v := strings.TrimSpace(q.Get("manufacturer")); v != "" {
			makeName = v
		}
		if v := strings.TrimSpace(q.Get("brand")); v != "" {
			makeName = v
		}
		if v := strings.TrimSpace(q.Get("model")); v != "" {
			model = v
		}
		if v := strings.TrimSpace(q.Get("year")); v != "" {
			year = v
		}
	}
	for _, cand := range serialCandidates(deviceID, desc, link) {
		if mapped, ok := modelFromSerial(cand); ok {
			model = mapped
			if makeName == "" {
				makeName = "reMarkable"
			}
			break
		}
	}
	if year == "" {
		if m := yearRegex.FindString(ls); m != "" {
			year = m
		}
	}
	return makeName, model, year
}

func serialCandidates(deviceID, desc, link string) []string {
	out := make([]string, 0, 8)
	push := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		out = append(out, s)
	}
	push(deviceID)
	if u, err := url.Parse(link); err == nil {
		q := u.Query()
		for _, k := range []string{"serial", "serialNumber", "deviceSerial", "sn"} {
			push(q.Get(k))
		}
	}
	for _, m := range serialLikeRegex.FindAllString(strings.ToUpper(desc+" "+link), -1) {
		push(m)
	}
	return out
}

func modelFromSerial(serial string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(serial))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	prefix6 := s
	if len(prefix6) > 6 {
		prefix6 = prefix6[:6]
	}
	if len(prefix6) >= 5 {
		key := prefix6[:5]
		switch key {
		case "RM02A":
			return "reMarkable Paper Pro", true
		case "RM03A":
			return "reMarkable Paper Pro Move", true
		case "RM110":
			return "reMarkable 2", true
		case "RM102":
			return "reMarkable 1", true
		case "RM12A":
			return "TBA", true
		}
	}
	return "", false
}
