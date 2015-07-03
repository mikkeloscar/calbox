package ical

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"time"
)

const (
	eol     = "\r\n"
	timeFmt = "20060102T150400Z"
	domain  = "@cal.moscar.net"
	prodid  = "calbox"
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
	buffer.WriteString("PRODID:-" + prodid)
	buffer.WriteString(eol)

	for _, e := range c.VEvents {
		buffer.WriteString(e.String())
	}

	buffer.WriteString("END:VCALENDAR")
	buffer.WriteString(eol)

	return buffer.String()
}

type VEvent struct {
	Summary     string
	Description string
	// Organizer string
	Start time.Time
	End   time.Time
}

func (e *VEvent) String() string {
	var buffer bytes.Buffer

	// create uid from summary and startdate
	h := sha1.New()
	io.WriteString(h, e.Summary+e.Start.Format(timeFmt))
	uid := base64.URLEncoding.EncodeToString(h.Sum(nil))

	buffer.WriteString("BEGIN:VEVENT")
	buffer.WriteString(eol)
	buffer.WriteString("UID:")
	buffer.WriteString(uid + domain)
	buffer.WriteString(eol)
	buffer.WriteString("DTSTAMP:")
	buffer.WriteString(time.Now().Format(timeFmt))
	buffer.WriteString(eol)
	buffer.WriteString("DTSTART:")
	buffer.WriteString(e.Start.Format(timeFmt))
	buffer.WriteString(eol)
	buffer.WriteString("DTEND:")
	buffer.WriteString(e.End.Format(timeFmt))
	buffer.WriteString(eol)
	buffer.WriteString("SUMMARY:")
	buffer.WriteString(e.Summary)
	buffer.WriteString(eol)

	if len(e.Description) > 0 {
		buffer.WriteString("DESCRIPTION:")
		buffer.WriteString(e.Description)
		buffer.WriteString(eol)
	}

	buffer.WriteString("END:VEVENT")
	buffer.WriteString(eol)

	return buffer.String()
}
