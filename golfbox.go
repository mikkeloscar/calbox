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

// get page myTimes.asp at golfbox.dk
func getTimes(username, password string) string {
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

	fmt.Printf("%s", string(body))
	r, _ := regexp.Compile(`(<\/?[^>]+(>|$)|\t)`)
	fmt.Printf("%s", r.ReplaceAll(body, []byte{}))

	return ""
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
	// login("14-1644", "2428")
	getTimes("14-1644", "2428")

}

func parseTimePage(page string) []*TeeTime {

	lines := strings.Split(page, "\n")

	for i, l := range lines {
		if len(l) > 9 && l[0:9] == "Mine tider" {
			return parseTimes(lines[i:])
		}
	}

	return nil
}

func parseTimes(lines []string) []*TeeTime {
	var teeTimes []*TeeTime

	for i, l := range lines {
		if start := matchingKey(l, "Klub: "); start > 0 {
			teeTime := &TeeTime{club: l[start : len(l)-1]}
			parseTime(lines[i:], teeTime)
			teeTimes = append(teeTimes, teeTime)
		}
	}

	return teeTimes
}

func parseTime(lines []string, tee *TeeTime) {
	var date string

	for i, l := range lines {
		if start := matchingKey(l, "Dato: "); start > 0 {
			date = l[start : len(l)-1]
			continue
		}

		if start := matchingKey(l, "Kl.: "); start > 0 {
			date += " " + l[start:len(l)-1]
			t, _ := time.Parse("02-01-06 15:04", date)
			tee.time = t
			continue
		}

		if len(l) > 1 && '0' < l[1] && l[1] < '9' {
			tee.players = parsePlayers
		}
	}
}

func matchingKey(line, key string) int {
	length := len(key)

	if len(line) > length && line[0:length] == key {
		return length + 1
	}

	return 0
}

type TeeTime struct {
	club    string
	time    time.Time
	players []*player
}

type player struct {
	name   string
	number string
	club   string
	hcp    string
}
