package messages

type CalendarEventsResponse struct {
	RetrievedAtUtc string          `json:"retrievedAtUtc"`
	Events         []CalendarEvent `json:"events"`
}

type CalendarEvent struct {
	ID                  string           `json:"id"`
	Title               string           `json:"title"`
	StartTimeUtc        string           `json:"startTimeUtc"`
	EndTimeUtc          string           `json:"endTimeUtc"`
	LastModifiedTimeUtc string           `json:"lastModifiedTimeUtc"`
	StartTime           string           `json:"startTime"`
	EndTime             string           `json:"endTime"`
	ResponseStatus      string           `json:"responseStatus"`
	Organizer           *CalendarPerson  `json:"organizer,omitempty"`
	Attendees           []CalendarPerson `json:"attendees,omitempty"`
}

type CalendarPerson struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
