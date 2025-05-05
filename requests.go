package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/log"
)

type requestHandler struct {
	jar *cookiejar.Jar
}

func newRequestHandler() requestHandler {
	jar, _ := cookiejar.New(nil)
	return requestHandler{
		jar: jar,
	}
}

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
func (p player) Description() string { return "Team: " + p.Team }
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
	Type    string   `json:"type"`
	items   []list.Item
	teams   bool
}

type newGame struct {
	GameId string `json:"gameId"`
}

type mySeat struct {
	Seat int `json:"seat"`
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

func (m requestHandler) statusRequest() bool {
	requestURL := fmt.Sprintf("%s/status", env.Server)
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
func (m requestHandler) registerRequest(register register) bool {
	payload, _ := json.Marshal(register)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/register", env.Server)

	client := &http.Client{
		Jar: m.jar,
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

func (m requestHandler) lobbyRequest() []list.Item {
	requestURL := fmt.Sprintf("%s/lobby", env.Server)
	items := []list.Item{}

	client := &http.Client{
		Jar: m.jar,
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

func (m requestHandler) makeGameRequest(gc gameConfig) newGame {
	payload, _ := json.Marshal(gc)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/makegame", env.Server)
	game := newGame{}

	client := &http.Client{Jar: m.jar}

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

func (m requestHandler) waitingRoomRequest() waitingRoom {
	requestURL := fmt.Sprintf("%s/waitingroom", env.Server)
	items := []list.Item{}

	client := &http.Client{
		Jar: m.jar,
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

	if waitingroom.Fill[2] == '4' && waitingroom.Type != "solo" {
		waitingroom.teams = true
	}

	return waitingroom
}

func (m requestHandler) leaveGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/leavegame", env.Server)

	client := &http.Client{
		Jar: m.jar,
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

func (m requestHandler) readyRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/ready", env.Server)

	client := &http.Client{Jar: m.jar}

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

func (m requestHandler) startGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/startgame", env.Server)

	client := &http.Client{Jar: m.jar}

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

func (m requestHandler) joinGameRequest(gameId gameId) bool {
	payload, _ := json.Marshal(gameId)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/joingame", env.Server)

	client := &http.Client{
		Jar: m.jar,
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

func (m requestHandler) handRequest() []card {
	requestURL := fmt.Sprintf("%s/hand", env.Server)
	var hand []card

	client := &http.Client{
		Jar: m.jar,
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
		hand = append(hand, newCard(handStrings[i]))
	}
	return hand
}

func (m requestHandler) playCardRequest(index handIndex) bool {
	payload, _ := json.Marshal(index)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/playcard", env.Server)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		// log.Errorf("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func (m requestHandler) actionsRequest() []action {
	var actions []action
	requestURL := fmt.Sprintf("%s/actions", env.Server)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error("error making http request: ", err)
		return actions
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: ", res.StatusCode)
		return actions
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return actions
	}

	json.Unmarshal([]byte(body.String()), &actions)

	return actions
}

func (m requestHandler) mySeatRequest() mySeat {
	var seat mySeat
	requestURL := fmt.Sprintf("%s/seat", env.Server)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error("error making http request: ", err)
		return seat
	}

	if res.StatusCode != http.StatusOK {
		return seat
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return seat
	}

	json.Unmarshal([]byte(body.String()), &seat)

	return seat
}

func (m requestHandler) changeTeamRequest(spectator bool) bool {
	reader := bytes.NewReader([]byte{})
	if spectator {
		reader = bytes.NewReader([]byte("{team:S}")) // Not worth implementing the JSON.
	}

	requestURL := fmt.Sprintf("%s/changeteam", env.Server)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		// log.Errorf("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func (m requestHandler) swapBottomCardRequest() bool {
	reader := bytes.NewReader([]byte{})

	requestURL := fmt.Sprintf("%s/swapBottomCard", env.Server)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request:", "err", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		// log.Errorf("bad status making http request:", "res.StatusCode", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func (m requestHandler) replayRequest(gameId gameId) []action {
	requestURL := fmt.Sprintf("%s/replay?gameId=%s", env.Server, gameId.GameId)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error("error making http request: ", err)
		return nil
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: ", res.StatusCode)
		return nil
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return nil
	}

	var actions []action
	json.Unmarshal([]byte(body.String()), &actions)

	return actions
}
