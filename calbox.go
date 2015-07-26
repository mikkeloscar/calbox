package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mikkeloscar/calbox/golfbox"
	"github.com/mikkeloscar/calbox/ical"
)

const (
	url    = "https://calbox.moscar.net/%s.ics"
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
	guid := vars["guid"]

	cal, _ := getCal(guid)
	w.Header().Set("Content-Type", "text/calendar; charset=UTF-8")
	w.Write(cal)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	pass := vars["pass"]

	guid, err := golfbox.Auth(user, pass)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid user\n"))
	} else {
		w.Write([]byte(fmt.Sprintf(url, guid)))
	}
}

func webServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/{guid}.ics", calHandler).Methods("GET")
	r.HandleFunc("/{user}/{pass}", authHandler).Methods("GET")

	http.Handle("/", r)

	log.Printf("Serving webserver on http://localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func getCal(guid string) ([]byte, error) {
	gb := golfbox.Conn(guid)

	times, err := gb.TeeTimes()
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
