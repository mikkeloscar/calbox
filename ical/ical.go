package ical

import (
	"bytes"
	"time"
)

const (
	eol = "\r\n"
)

type VCalendar struct {
	VEvents []*VEvent
}

func (c *VCalendar) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("BEGIN:VCALENDAR")
	buffer.WriteString(eol)
	buffer.WriteString("VERSION:2.0")
	buffer.WriteString(eol)

	for _, e := range c.VEvents {
		buffer.WriteString(e.String())
	}

	buffer.WriteString("END:VCALENDAR")
	buffer.WriteString(eol)

	return buffer.String()
}

type VEvent struct {
	uid       string
	Summary   string
	Organizer string
	DTStamp   time.Time
	DTStart   time.Time
	DTEnd     time.Time
}

func (e *VEvent) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("BEGIN:EVENT")
	buffer.WriteString(eol)
	// buffer.WriteString(e.DTStamp)
	buffer.WriteString("END:EVENT")
	buffer.WriteString(eol)

	return buffer.String()
}
