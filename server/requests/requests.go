package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"brisca.sh/server/game"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type ServerType int

const (
	BROWSER ServerType = iota
	GAME    ServerType = iota
)

type Handler struct {
	jar           http.CookieJar
	GameServer    string
	BrowserServer string
	RefreshBy     time.Time
}

func NewHandler(browser string, gameServer string) Handler {
	return Handler{
		jar:           &SharedCookieJar{CookieSlice: []*http.Cookie{}},
		BrowserServer: browser,
		GameServer:    gameServer,
	}
}

type RefreshByMsg time.Time

func (m *Handler) RefreshSessionCheck(before time.Duration) tea.Msg {
	var rtn tea.Msg

	now := time.Now().UTC()
	expires := m.RefreshBy.UTC()
	timeBefore := expires.Add(-before)
	if !(now.After(timeBefore) && now.Before(expires)) {
		return rtn
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debug("Time: ", "timeBefore", timeBefore)
		log.Debug("Time: ", "now       ", now)
		log.Debug("Time: ", "expires   ", expires)
	}

	rtn = m.refreshSessionRequest()

	return rtn
}

type Register struct {
	Username string `json:"username"`
}

type Game struct {
	GameId string `json:"gameId"`
	Fill   string `json:"fill"`
	Server string `json:"server"`
}

func (g Game) Title() string       { return g.GameId }
func (g Game) Description() string { return "Fill: " + g.Fill }
func (g Game) FilterValue() string { return g.GameId }

type GameConfig struct {
	GameType       string `json:"gameType"`
	MaxPlayers     int    `json:"maxPlayers"`
	SwapBottomCard bool   `json:"swapBottomCard"`
}

type GamesList struct {
	Games []Game `json:"games"`
}

type Player struct {
	Ready bool   `json:"ready"`
	Name  string `json:"name"`
	Team  string `json:"team"` // Only relevant for 4 player games.
}

func (p Player) Title() string       { return p.Name + ": " + p.ReadyString() }
func (p Player) Description() string { return "Team: " + p.Team }
func (p Player) FilterValue() string { return p.Name }
func (p Player) ReadyString() string {
	if p.Ready {
		return "ready"
	} else {
		return "not ready"
	}
}

type NewGame struct {
	GameId     string `json:"gameId"`
	GameServer string `json:"gameServer"`
}

type MySeat struct {
	Seat int `json:"seat"`
}

// {"port":"9004","host":"browser","expiration":"2025-05-21T23:13:58.220082540Z"}
type Lease struct {
	Port          string `json:"port"`
	Host          string `json:"host"`
	ExpirationStr string `json:"expiration"`
	Expiration    time.Time
}

type WaitingRoom struct {
	Players  []Player `json:"players"`
	Fill     string   `json:"fill"`
	Started  bool     `json:"started"`
	TimedOut bool     `json:"timedOut"`
	Type     string   `json:"type"`
	Items    []list.Item
	Teams    bool
}

func (wr WaitingRoom) String() string {
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
		sb.WriteString(wr.Players[i].ReadyString())
		sb.WriteString(" name:")
		sb.WriteString(wr.Players[i].Name)
		sb.WriteString(" team:")
		sb.WriteString(wr.Players[i].Team)
	}
	return sb.String()
}

type GameId struct {
	GameId string `json:"gameId"`
}

type handIndex struct {
	Index int `json:"index"`
}

func (m Handler) StatusRequest(stype ServerType) bool {
	var url string
	if stype == BROWSER {
		url = m.BrowserServer
	} else {
		url = m.GameServer
	}

	requestURL := fmt.Sprintf("%s/status", url)
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

func (m *Handler) RegisterRequest(register Register, server string) bool {
	payload, _ := json.Marshal(register)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/register", server)

	client := &http.Client{
		Jar: m.jar,
	}

	tryString := fmt.Sprintf("%s, %s", requestURL, register.Username)
	log.Debug("Register, trying: ", "tryString", tryString)

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: %s\n", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return false
	}

	log.Debug("Register success:", "register.Username", register.Username)

	cookies := res.Cookies()
	for i := range cookies {
		cookie := cookies[i]
		if cookie.Name != "userId" {
			continue
		}
		log.Debug("Inspecting cookie expiration:", "cookie.RawExpires", cookie.RawExpires)
		m.RefreshBy, err = time.Parse(time.RFC1123, cookie.RawExpires)
		if err != nil {
			log.Error("Error parsing time by RFC1123 failed.", "cookie.RawExpires", cookie.RawExpires)
			log.Error("Error:", "err", err)
			return false
		}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func (m Handler) LobbyRequest() []list.Item {
	requestURL := fmt.Sprintf("%s/lobby", m.BrowserServer)
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

	games := GamesList{}
	json.Unmarshal([]byte(body.String()), &games)

	for i := range games.Games {
		items = append(items, games.Games[i])
	}

	return items
}

func (m *Handler) MakeGameRequest(gc GameConfig) NewGame {
	payload, _ := json.Marshal(gc)
	reader := bytes.NewReader(payload)
	game := NewGame{}

	requestURL := fmt.Sprintf("%s/makeGame", m.GameServer)

	client := &http.Client{Jar: m.jar}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: ", "err", err.Error())
		return game
	}

	if res.StatusCode != http.StatusOK {
		log.Error("While making http request:", "requestURL", requestURL, "res", res)
		return game
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error("error making http request: ", "err", err)
		return game
	}

	json.Unmarshal([]byte(body.String()), &game)

	m.GameServer = fmt.Sprintf("http://%s", game.GameServer)

	log.Debug("Configured Server: ", "GameServer", m.GameServer, "now", time.Now())

	return game
}

func (m Handler) WaitingRoomRequest() WaitingRoom {
	requestURL := fmt.Sprintf("%s/waitingroom", m.GameServer)
	items := []list.Item{}

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return WaitingRoom{}
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: %d\n", res.StatusCode)
		return WaitingRoom{}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error making http request: %s\n", err.Error()))
		return WaitingRoom{}
	}

	waitingroom := WaitingRoom{}
	json.Unmarshal([]byte(body.String()), &waitingroom)

	for i := range waitingroom.Players {
		items = append(items, waitingroom.Players[i])
	}

	waitingroom.Items = items

	if waitingroom.Fill[2] == '4' && waitingroom.Type != "solo" {
		waitingroom.Teams = true
	}

	return waitingroom
}

func (m Handler) LeaveGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/leavegame", m.GameServer)

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

func (m Handler) ReadyRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/ready", m.GameServer)

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

func (m Handler) StartGameRequest() bool {
	reader := bytes.NewBufferString("")
	requestURL := fmt.Sprintf("%s/startgame", m.GameServer)

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

func (m *Handler) JoinGameRequest(gameId GameId, server, username string) bool {
	payload, _ := json.Marshal(gameId)
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("http://%s/joingame", server)

	tmpGameServer := "http://" + server

	reg := Register{
		Username: username,
	}

	success := m.RegisterRequest(reg, tmpGameServer)
	if !success {
		log.Error("Error registering to gameServer: ", "tmpGameServer", tmpGameServer)
		return false
	}

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: ", "requestURL", requestURL, "err", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: ", "requestURL", requestURL, "StatusCode", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	m.GameServer = tmpGameServer

	return true
}

func (m *Handler) JoinPrivateGameRequest(gameId GameId, username string) bool {
	requestURL := fmt.Sprintf("%s/joinprivategame?gameId=%s", m.BrowserServer, gameId.GameId)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		log.Error("error making http request: ", "requestURL", requestURL, "err", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: ", "requestURL", requestURL, "StatusCode", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	var game Game

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		log.Error("error making http request: ", "err", err)
		return false
	}

	json.Unmarshal([]byte(body.String()), &game)

	tmpGameServer := "http://" + game.Server

	reg := Register{
		Username: username,
	}

	success := m.RegisterRequest(reg, tmpGameServer)
	if !success {
		log.Error("Error registering to gameServer: ", "tmpGameServer", tmpGameServer)
		return false
	}

	payload, _ := json.Marshal(game)
	reader := bytes.NewReader(payload)
	requestURL = fmt.Sprintf("%s/joingame", tmpGameServer)

	client = &http.Client{
		Jar: m.jar,
	}

	res, err = client.Post(requestURL, "raw", reader)
	if err != nil {
		log.Error("error making http request: ", "requestURL", requestURL, "err", err)
		return false
	}

	if res.StatusCode != http.StatusOK {
		log.Error("bad status making http request: ", "requestURL", requestURL, "StatusCode", res.StatusCode)
		return false
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	m.GameServer = tmpGameServer

	return true
}

func (m Handler) HandRequest() []game.Card {
	requestURL := fmt.Sprintf("%s/hand", m.GameServer)
	var hand []game.Card

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

func handFromStrings(handStrings []string) []game.Card {
	var hand []game.Card
	for i := range len(handStrings) {
		hand = append(hand, game.NewCard(handStrings[i]))
	}
	return hand
}

func (m Handler) PlayCardRequest(index int) bool {
	payload, _ := json.Marshal(handIndex{Index: index})
	reader := bytes.NewReader(payload)
	requestURL := fmt.Sprintf("%s/playcard", m.GameServer)

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

func (m Handler) ActionsRequest() []game.Action {
	var actions []game.Action
	requestURL := fmt.Sprintf("%s/actions", m.GameServer)

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

func (m Handler) MySeatRequest() MySeat {
	var seat MySeat
	requestURL := fmt.Sprintf("%s/seat", m.GameServer)

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

func (m Handler) ChangeTeamRequest(spectator bool) bool {
	reader := bytes.NewReader([]byte{})
	if spectator {
		reader = bytes.NewReader([]byte("{team:S}")) // Not worth implementing the JSON.
	}

	requestURL := fmt.Sprintf("%s/changeteam", m.GameServer)

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

func (m Handler) SwapBottomCardRequest() bool {
	reader := bytes.NewReader([]byte{})

	requestURL := fmt.Sprintf("%s/swapBottomCard", m.GameServer)

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

func (m Handler) ReplayRequest(gameId GameId) []game.Action {
	requestURL := fmt.Sprintf("%s/replay?gameId=%s", m.BrowserServer, gameId.GameId)

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

	var actions []game.Action
	json.Unmarshal([]byte(body.String()), &actions)

	return actions
}

func (m *Handler) refreshSessionRequest() tea.Msg {
	var msg RefreshByMsg
	requestURL := fmt.Sprintf("%s/refresh", m.BrowserServer)

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

	cookies := res.Cookies()
	for i := range cookies {
		cookie := cookies[i]
		if cookie.Name != "userId" {
			continue
		}
		msg = RefreshByMsg(cookie.Expires)
		log.Debug("Cookies from body: ", "cookie", cookie)
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return msg
}
