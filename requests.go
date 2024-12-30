package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/log"
)

var (
	jar, _             = cookiejar.New(nil)
	baseurl            = "http://localhost:8000"
	CARD_NUMBER_INDEX  = 0
	CARD_VALUE_INDEX   = 1
	CARD_SCORE_INDEX   = 2
	CARDS_WITHOUT_SKIP = [][]int{
		{1, 12, 11},
		{2, 1, 0},
		{3, 11, 10},
		{4, 2, 0},
		{5, 3, 0},
		{6, 4, 0},
		{7, 5, 0},
		{8, 6, 0},
		{9, 7, 0},
		{10, 8, 2},
		{11, 9, 3},
		{12, 10, 4},
	}
)

type register struct {
	Username string `json:"username"`
}

type game struct {
	GameId string `json:"gameId"`
	Fill   string `json:"fill"`
}

func (g game) Title() string       { return g.GameId }
func (g game) Description() string { return "Fill: " + g.Fill }
func (g game) FilterValue() string { return g.GameId }

type gameConfig struct {
	GameType       string `json:"gameType"`
	MaxPlayers     int    `json:"maxPlayers"`
	SwapBottomCard bool   `json:"swapBottomCard"`
}

type gamesList struct {
	Games []game `json:"games"`
}

type player struct {
	Ready bool   `json:"ready"`
	Name  string `json:"name"`
	Team  string `json:"team"` // Only relevant for 4 player games.
}

func (p player) Title() string       { return p.Name + ": " + p.ready() }
func (p player) Description() string { return p.Team }
func (p player) FilterValue() string { return p.Name }
func (p player) ready() string {
	if p.Ready {
		return "ready"
	} else {
		return "not ready"
	}
}

type waitingRoom struct {
	Players []player `json:"players"`
	Fill    string   `json:"fill"`
	Started bool     `json:"started"`
	items   []list.Item
}

type newGame struct {
	GameId string `json:"gameId"`
}

func (wr waitingRoom) String() string {
	sb := strings.Builder{}
	sb.WriteString("fill:")
	sb.WriteString(wr.Fill)
	sb.WriteString(" started:")
	if wr.Started {
		sb.WriteString("true")
	} else {
		sb.WriteString("false")
	}
	for i := range wr.Players {
		sb.WriteString(" ready:")
		sb.WriteString(wr.Players[i].ready())
		sb.WriteString(" name:")
		sb.WriteString(wr.Players[i].Name)
		sb.WriteString(" team:")
		sb.WriteString(wr.Players[i].Team)
	}
	return sb.String()
}

type gameId struct {
	GameId string `json:"gameId"`
}

type handIndex struct {
	Index int `json:"index"`
}

// Not JSON types
type card struct {
	suit  string
	num   int
	val   int
	score int
}

func statusRequest() bool {
	requestURL := fmt.Sprintf("%s/status", baseurl)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	return true
}

// {"username" : "Guest"}
func registerRequest(register register) bool {
	payload, _ := json.Marshal(register)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/register", baseurl)

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func lobbyRequest() []list.Item {
	requestURL := fmt.Sprintf("%s/lobby", baseurl)
	items := []list.Item{}

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return items
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return items
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return items
	}

	games := gamesList{}
	json.Unmarshal([]byte(body.String()), &games)

	for i := range games.Games {
		items = append(items, games.Games[i])
	}

	return items
}

func makeGameRequest(gc gameConfig) newGame {
	payload, _ := json.Marshal(gc)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/makegame", baseurl)
	game := newGame{}

	client := &http.Client{Jar: jar}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return game
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return game
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return game
	}

	json.Unmarshal([]byte(body.String()), &game)

	return game
}

func waitingRoomRequest() waitingRoom {
	requestURL := fmt.Sprintf("%s/waitingroom", baseurl)
	items := []list.Item{}

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return waitingRoom{}
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return waitingRoom{}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return waitingRoom{}
	}

	waitingroom := waitingRoom{}
	json.Unmarshal([]byte(body.String()), &waitingroom)

	for i := range waitingroom.Players {
		items = append(items, waitingroom.Players[i])
	}

	waitingroom.items = items

	return waitingroom
}

func leaveGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/leavegame", baseurl)

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func readyRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/ready", baseurl)

	client := &http.Client{Jar: jar}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func startGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/startgame", baseurl)

	client := &http.Client{Jar: jar}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func joinGameRequest(gameId gameId) bool {
	payload, _ := json.Marshal(gameId)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/joingame", baseurl)

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func handRequest() []card {
	requestURL := fmt.Sprintf("%s/hand", baseurl)
	var hand []card

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return hand
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return hand
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return hand
	}

	var handStrings []string
	json.Unmarshal([]byte(body.String()), &handStrings)

	hand = handFromStrings(handStrings)

	return hand
}

func handFromStrings(handStrings []string) []card {
	var hand []card
	for i := range len(handStrings) {
		var card card

		halves := strings.Split(handStrings[i], ":")
		suitString := halves[0]
		num, err := strconv.Atoi(halves[1])
		if err != nil {
			log.Error("Error parsing str to int for hand request.")
			return hand
		}
		index := num - 1

	suitSwitch:
		switch suitString {
		case "ORO":
			card.suit = "🪙"
			break suitSwitch
		case "COPA":
			card.suit = "🏆"
			break suitSwitch
		case "BASTO":
			card.suit = "🪵"
			break suitSwitch
		case "ESPADA":
			card.suit = "⚔️"
			break suitSwitch
		}

		card.num = CARDS_WITHOUT_SKIP[index][CARD_NUMBER_INDEX]
		card.val = CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX]
		card.score = CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX]

		hand = append(hand, card)
	}
	return hand
}

func playCardRequest(index handIndex) bool {
	payload, _ := json.Marshal(index)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/playcard", baseurl)

	client := &http.Client{
		Jar: jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}
