package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mikkeloscar/calbox/golfbox"
	"github.com/mikkeloscar/calbox/ical"
)

const (
	domain = "@cal.moscar.net"
	prodid = "calbox"
)

func main() {
	// times := golfbox.GetTimes("14-1644", "2428")
	// for _, t := range times {
	// 	fmt.Println(t.Club)
	// 	fmt.Println(t.Time)
	// 	for _, p := range t.Players {
	// 		fmt.Printf("%#v\n", p)
	// 	}
	// }

	port := flag.Int("p", 4220, "HTTP server port")
	flag.Parse()

	webServer(*port)
}

func calHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	user := q.Get("user")
	pass := q.Get("pass")

	if len(user) == 0 || len(pass) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid user\n"))
	} else {
		// TODO handle err
		cal, _ := getCal(user, pass)
		io.WriteString(w, cal)
	}
}

func webServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", calHandler).Methods("GET")

	http.Handle("/", r)

	log.Printf("Serving webserver on http://localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func getCal(user, pass string) (string, error) {
	gb := golfbox.Conn(user, pass)

	times, err := gb.GetTimes()
	if err != nil {
		return "", err
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
		Domain:  domain,
		ProdID:  prodid,
		VEvents: events,
	}

	return cal.String(), nil
}
