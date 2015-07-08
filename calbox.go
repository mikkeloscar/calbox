package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mikkeloscar/calbox/golfbox"
	"github.com/mikkeloscar/calbox/ical"
)

const (
	domain = "@calbox.moscar.net"
	prodid = "-//moscar.net/Calbox//DA"
)

func main() {
	port := flag.Int("p", 4220, "HTTP server port")
	flag.Parse()

	webServer(*port)
}

func calHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO handle err
	user, pass, _ := decodeID(id)

	if len(user) == 0 || len(pass) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid user\n"))
	} else {
		// TODO handle err
		cal, _ := getCal(user, pass)
		w.Header().Set("Content-Type", "text/calendar; charset=UTF-8")
		w.Write(cal)
	}
}

func webServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/{id}.ics", calHandler).Methods("GET")

	http.Handle("/", r)

	log.Printf("Serving webserver on http://localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func getCal(user, pass string) ([]byte, error) {
	gb := golfbox.Conn(user, pass)

	times, err := gb.GetTimes()
	if err != nil {
		return []byte{}, err
	}

	events := make([]*ical.VEvent, 0, len(times))

	for _, teeTime := range times {
		// new event
		event := &ical.VEvent{
			Summary:     "Golf tid: " + teeTime.Club,
			Description: teeTime.StrPlayers(),
			Start:       *teeTime.Time,
			End:         teeTime.Time.Add(time.Duration(4) * time.Hour),
		}

		events = append(events, event)
	}

	cal := &ical.VCalendar{
		Name:    "Calbox",
		Domain:  domain,
		ProdID:  prodid,
		VEvents: events,
	}

	return []byte(cal.String()), nil
}

func decodeID(id string) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return "", "", err
	}

	split := strings.Split(string(data), ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid id format")
	}

	return split[0], split[1], nil
}
