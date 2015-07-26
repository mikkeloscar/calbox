package golfbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	baseURL     = "http://restapi.golfbox.dk"
	authURL     = "/GolferPortal/Member/DK/Security/%s/%s/Authenticate"
	teeTimesURL = "/GolferPortal/GolfBox/App/Booking/TeeTimes/%s"

	APIUser     = "GolferPortal"
	APIPassword = "CSaHWV6jEV"
)

var stripHTMLregex = regexp.MustCompile(`(<\/?[^>]+(>|$)|\t)`)

type GolfBox struct {
	username string
	password string
	guid     string
}

type AuthResp struct {
	Guid string
}

type TeeTimesResp []TeeTimeResp

type TeeTimeResp struct {
	ClubName string `json:"Club_Name"`
	TeeTime  string
	Members  []MemberResp `json:"members"`
}

type MemberResp struct {
	ClubName  string `json:"Club_Name"`
	ClubID    string `json:"MemberClub_ID"`
	HCP       string `json:"Member_HCP"`
	Firstname string `json:"Member_Firstname"`
	Lastname  string `json:"Member_Lastname"`
}

func Conn(username, password string) (*GolfBox, error) {
	req, err := http.NewRequest("GET", baseURL+fmt.Sprintf(authURL, username, password), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(APIUser, APIPassword)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	result := new(AuthResp)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
	}

	return &GolfBox{username, password, result.Guid}, nil
}

// TeeTimes API call
func (gb *GolfBox) TeeTimes() ([]*TeeTime, error) {
	req, err := http.NewRequest("GET", baseURL+fmt.Sprintf(teeTimesURL, gb.guid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(APIUser, APIPassword)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	result := new(TeeTimesResp)
	err = dec.Decode(result)
	if err != nil {
		return nil, err
	}

	return parseTeeTimeResp(result)
}

func parseTeeTimeResp(respTeeTimes *TeeTimesResp) ([]*TeeTime, error) {
	teeTimes := make([]*TeeTime, 0, len(*respTeeTimes))

	for _, teeTime := range *respTeeTimes {
		// parse date
		loc, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			panic(err)
		}

		date, err := time.ParseInLocation("20060102T150405", teeTime.TeeTime, loc)
		if err != nil {
			return nil, err
		}

		// parse members
		players, err := parseMembersResp(teeTime.Members)
		if err != nil {
			return nil, err
		}

		tee := &TeeTime{
			Club:    teeTime.ClubName,
			Time:    &date,
			Players: players,
		}

		teeTimes = append(teeTimes, tee)
	}

	return teeTimes, nil
}

func parseMembersResp(respMembers []MemberResp) ([]*Player, error) {
	players := make([]*Player, 0, len(respMembers))

	for _, p := range respMembers {
		val, err := strconv.ParseInt(p.HCP, 10, 0)
		if err != nil {
			return nil, err
		}

		player := &Player{
			Name:   fmt.Sprintf("%s %s", p.Firstname, p.Lastname),
			Number: p.ClubID,
			Club:   p.ClubName,
			HCP:    float32(val) / 10000.0,
		}

		players = append(players, player)
	}

	return players, nil
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
			buf.WriteString("\\n\\n")
		}
	}

	return buf.String()
}

type Player struct {
	Name   string
	Number string
	Club   string
	HCP    float32
}

func (p *Player) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s (%.1f)\\n", p.Name, p.HCP))
	buf.WriteString(fmt.Sprintf("%s (%s)", p.Number, p.Club))

	return buf.String()
}
