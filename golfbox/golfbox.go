package golfbox

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

type GolfBox struct {
	username string
	password string
}

// Conn setup connection to golfbox (username and password)
func Conn(username, password string) *GolfBox {
	return &GolfBox{username, password}
}

// GetTimes get page myTimes.asp at golfbox.dk
func (gb *GolfBox) GetTimes() ([]*TeeTime, error) {
	cookies := gb.login()

	req, err := http.NewRequest("GET", baseURL+myTimesURL, nil)
	if err != nil {
		return nil, err
	}

	// add cookies
	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseTimePage(string(stripHTMLregex.ReplaceAll(body, []byte{}))), nil
}

// login to golfbox.dk with username and password. Return session cookie.
func (gb *GolfBox) login() []*http.Cookie {
	var form bytes.Buffer

	form.WriteString("loginform.submitted=true&command=login")
	form.WriteString("&loginform.username=")
	form.WriteString(gb.username)
	form.WriteString("&loginform.password=")
	form.WriteString(gb.password)
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
			teeTime := &TeeTime{Club: p.current()[start:]}
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
			loc, err := time.LoadLocation("Europe/Copenhagen")
			if err != nil {
				panic(err)
			}

			date += " " + p.current()[start:]
			t, _ := time.ParseInLocation("02-01-06 15:04", date, loc)
			tee.Time = &t
			p.i++
			continue
		}

		if tee.Time != nil {
			tee.Players = p.parsePlayers()
			break
		}

		p.i++
	}
}

func (p *timeParser) parsePlayers() []*Player {
	var players []*Player

	for p.i < len(p.lines) {
		l := p.current()
		if len(l) == 1 && '0' < l[0] && l[0] < '9' {
			pl := &Player{}
			p.i++
			pl.Name = p.current()
			p.i++
			pl.Number = p.current()
			p.i++
			pl.Club = p.current()
			p.i++
			pl.HCP = p.current()
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
	Club    string
	Time    *time.Time
	Players []*Player
}

func (t *TeeTime) StrPlayers() string {
	var buf bytes.Buffer

	for i, p := range t.Players {
		buf.WriteString(p.String())

		if i < len(t.Players)-1 {
			buf.WriteByte('\n')
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

type Player struct {
	Name   string
	Number string
	Club   string
	HCP    string
}

func (p *Player) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s (%s)\n", p.Name, p.Number))
	buf.WriteString(fmt.Sprintf("%s (%s)", p.Number, p.Club))

	return buf.String()
}
