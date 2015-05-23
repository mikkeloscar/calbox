package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	baseURL    = "http://www.golfbox.dk"
	loginURL   = "/login.asp"
	myTimesURL = "/site/my_golfBox/myTimes.asp"
)

var stripHTMLregex = regexp.MustCompile(`(<\/?[^>]+(>|$)|\t)`)

// get page myTimes.asp at golfbox.dk
func getTimes(username, password string) []*TeeTime {
	cookies := login(username, password)

	req, err := http.NewRequest("GET", baseURL+myTimesURL, nil)
	if err != nil {
		panic(err)
	}

	// add cookies
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", body)

	return parseTimePage(string(stripHTMLregex.ReplaceAll(body, []byte{})))
}

// login to golfbox.dk with username and password. Return session cookie.
func login(username, password string) []*http.Cookie {
	var form bytes.Buffer

	form.WriteString("loginform.submitted=true&command=login")
	form.WriteString("&loginform.username=")
	form.WriteString(username)
	form.WriteString("&loginform.password=")
	form.WriteString(password)
	form.WriteString("&LOGIN=LOGIN")

	req, err := http.NewRequest("POST", baseURL+loginURL, &form)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Transport{}
	resp, err := client.RoundTrip(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return resp.Cookies()
}

func main() {
	getTimes("14-1644", "2428")

}

// functions for parsing mytimes.asp HTML page
type timeParser struct {
	lines []string
	i     int
}

func (p *timeParser) current() string {
	return p.lines[p.i]
}

func parseTimePage(page string) []*TeeTime {

	var p timeParser

	p.lines = strings.Split(page, "\r\n")

	for p.i < len(p.lines) {
		if p.lines[p.i] == "Mine tider" {
			return p.parseTimes()
		}
		p.i++
	}

	return nil
}

func (p *timeParser) parseTimes() []*TeeTime {
	var teeTimes []*TeeTime

	for p.i < len(p.lines) {
		if start := matchingKey(p.current(), "Klub: "); start > 0 {
			teeTime := &TeeTime{club: p.current()[start:]}
			p.i++
			p.parseTime(teeTime)
			teeTimes = append(teeTimes, teeTime)
			continue
		}

		p.i++
	}

	return teeTimes
}

func (p *timeParser) parseTime(tee *TeeTime) {
	var date string

	for p.i < len(p.lines) {
		if start := matchingKey(p.current(), "Dato: "); start > 0 {
			date = p.current()[start:]
			p.i++
			continue
		}

		if start := matchingKey(p.current(), "Kl.: "); start > 0 {
			date += " " + p.current()[start:]
			t, _ := time.Parse("02-01-06 15:04", date)
			tee.time = &t
			p.i++
			continue
		}

		if tee.time != nil {
			tee.players = p.parsePlayers()
			break
		}

		p.i++
	}
}

func (p *timeParser) parsePlayers() []*player {
	var players []*player

	for p.i < len(p.lines) {
		l := p.current()
		if len(l) == 1 && '0' < l[0] && l[0] < '9' {
			pl := &player{}
			p.i++
			pl.name = p.current()
			p.i++
			pl.number = p.current()
			p.i++
			pl.club = p.current()
			p.i++
			pl.hcp = p.current()
			players = append(players, pl)
		}

		if start := matchingKey(p.current(), "Klub: "); start > 0 {
			break
		}

		p.i++
	}

	return players
}

func matchingKey(line, key string) int {
	length := len(key)

	if len(line) > length && line[0:length] == key {
		return length
	}

	return 0
}

type TeeTime struct {
	club    string
	time    *time.Time
	players []*player
}

type player struct {
	name   string
	number string
	club   string
	hcp    string
}
