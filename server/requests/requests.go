package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"brisca.sh/server/game"
	"brisca.sh/server/logwrapper"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ServerType int

const (
	BROWSER ServerType = iota
	GAME    ServerType = iota
)

var (
	logger = logwrapper.NewLogger()
)

type Handler struct {
	jar           http.CookieJar
	gameServer    string
	browserServer string
	RefreshBy     time.Time
}

func NewHandler(browser string, gameServer string) Handler {
	return Handler{
		jar:           &SharedCookieJar{CookieSlice: []*http.Cookie{}},
		browserServer: browser,
		gameServer:    gameServer,
	}
}

func (m Handler) GameServer() string {
	return m.gameServer
}

func (m *Handler) SetGameServer(hostAndPort string) {
	m.gameServer = fmt.Sprintf("http://%s", hostAndPort)
}

func (m Handler) BroserServer() string {
	return m.browserServer
}

func (m *Handler) SetBrowserServer(hostAndPort string) {
	m.browserServer = fmt.Sprintf("http://%s", hostAndPort)
}

type HttpStatusErr struct {
	err error
	res *http.Response
}

func (m HttpStatusErr) Error() string {
	return m.err.Error()
}

func (m *Handler) getRequest(url string) (string, error) {

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(url)
	if err != nil {
		logger.Error("Get:", "url", url, "err", err, "res", res)
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", HttpStatusErr{
			err: fmt.Errorf("Bad Status: %d", res.StatusCode),
			res: res,
		}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		logger.Error("Get:", "url", url, "err", err, "res", res, "body", body.String())
		return "", err
	}

	return body.String(), nil

}

func (m *Handler) postRequest(url string, payload []byte) (string, error) {
	reader := bytes.NewReader(payload)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Post(url, "raw", reader)
	if err != nil {
		logger.Error("Get:", "url", url, "err", err, "res", res)
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", HttpStatusErr{
			err: fmt.Errorf("Bad Status: %d", res.StatusCode),
			res: res,
		}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	body := new(strings.Builder)
	_, err = io.Copy(body, res.Body)
	if err != nil {
		logger.Error("Get:", "url", url, "err", err, "res", res, "body", body.String())
		return "", err
	}

	return body.String(), nil

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

	if logger.Enabled(slog.LevelDebug) {
		logger.Debug("Time: ", "timeBefore", timeBefore)
		logger.Debug("Time: ", "now       ", now)
		logger.Debug("Time: ", "expires   ", expires)
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
		url = m.browserServer
	} else {
		url = m.gameServer
	}

	requestURL := fmt.Sprintf("%s/status", url)
	_, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("StatusRequest:", "url", requestURL, "err", err)
		return false
	}

	return true
}

func (m *Handler) RegisterRequest(register Register, server string) bool {
	payload, _ := json.Marshal(register)
	reader := bytes.NewReader(payload)
	if server == "" {
		server = m.browserServer
	}
	requestURL := fmt.Sprintf("%s/register", server)

	client := &http.Client{
		Jar: m.jar,
	}

	tryString := fmt.Sprintf("%s, %s", requestURL, register.Username)
	logger.Debug("Register, trying: ", "tryString", tryString)

	res, err := client.Post(requestURL, "raw", reader)
	if err != nil {
		logger.Error("RegisterRequest:", "url", requestURL, "err", err, "res", res)
		return false
	}

	if res.StatusCode != http.StatusOK {
		logger.Error("RegisterRequest:", "url", requestURL, "err", err, "res", res)
		return false
	}

	logger.Debug("Register success:", "register.Username", register.Username)

	cookies := res.Cookies()
	for i := range cookies {
		cookie := cookies[i]
		if cookie.Name != "userId" {
			continue
		}
		logger.Debug("Inspecting cookie expiration:", "cookie.RawExpires", cookie.RawExpires)
		m.RefreshBy, err = time.Parse(time.RFC1123, cookie.RawExpires)
		if err != nil {
			logger.Error("Error parsing time by RFC1123 failed.", "cookie.RawExpires", cookie.RawExpires)
			logger.Error("Error:", "err", err)
			return false
		}
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return true
}

func (m Handler) LobbyRequest() []list.Item {
	requestURL := fmt.Sprintf("%s/lobby", m.browserServer)
	items := []list.Item{}

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("Lobby Request:", "url", requestURL, "err", err, "body", body)
		return items
	}

	games := GamesList{}
	json.Unmarshal([]byte(body), &games)

	for i := range games.Games {
		items = append(items, games.Games[i])
	}

	return items
}

func (m *Handler) MakeGameRequest(gc GameConfig) NewGame {
	payload, _ := json.Marshal(gc)
	game := NewGame{}

	requestURL := fmt.Sprintf("%s/makeGame", m.gameServer)

	body, err := m.postRequest(requestURL, payload)
	if err != nil {
		logger.Error("MakeGameRequest:", "url", requestURL, "err", err, "body", body)
		return game
	}

	json.Unmarshal([]byte(body), &game)

	return game
}

func (m Handler) WaitingRoomRequest() WaitingRoom {
	requestURL := fmt.Sprintf("%s/waitingroom", m.gameServer)
	items := []list.Item{}

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("WaitingRoom:", "url", requestURL, "err", err, "body", body)
		return WaitingRoom{}
	}

	waitingroom := WaitingRoom{}
	json.Unmarshal([]byte(body), &waitingroom)

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
	requestURL := fmt.Sprintf("%s/leavegame", m.gameServer)

	body, err := m.postRequest(requestURL, []byte(""))
	if err != nil {
		// statusErr, ok := err.(HttpStatusErr)
		// if ok {
		// 	logger.Debug("HttpStatusErr type assertion ran.")
		// 	if statusErr.res.StatusCode >= 400 && 500 > statusErr.res.StatusCode {
		// 		return false
		// 	}
		// }
		logger.Error("LeaveGameRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m Handler) ReadyRequest() bool {
	requestURL := fmt.Sprintf("%s/ready", m.gameServer)

	body, err := m.postRequest(requestURL, []byte(""))
	if err != nil {
		logger.Error("ReadyRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m Handler) StartGameRequest() bool {
	requestURL := fmt.Sprintf("%s/startgame", m.gameServer)

	body, err := m.postRequest(requestURL, []byte(""))
	if err != nil {
		logger.Error("StartGameRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m *Handler) JoinGameRequest(gameId GameId, server, username string) NewGame {
	game := NewGame{}
	payload, _ := json.Marshal(gameId)
	requestURL := fmt.Sprintf("http://%s/joingame", server)

	tmpGameServer := "http://" + server

	reg := Register{
		Username: username,
	}

	success := m.RegisterRequest(reg, tmpGameServer)
	if !success {
		logger.Error("Error registering to gameServer: ", "tmpGameServer", tmpGameServer)
		return game
	}

	body, err := m.postRequest(requestURL, payload)
	if err != nil {
		logger.Error("JoinGameRequest:", "url", requestURL, "err", err, "body", body)
		return game
	}

	game.GameId = gameId.GameId
	game.GameServer = server

	return game
}

func (m *Handler) JoinPrivateGameRequest(gameId GameId, username string) NewGame {
	gameRtn := NewGame{}
	requestURL := fmt.Sprintf("%s/joinprivategame?gameId=%s", m.browserServer, gameId.GameId)

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("JoinPrivateGameRequest:", "url", requestURL, "err", err, "body", body)
		return gameRtn
	}

	var game Game

	json.Unmarshal([]byte(body), &game)

	tmpGameServer := "http://" + game.Server

	reg := Register{
		Username: username,
	}

	success := m.RegisterRequest(reg, tmpGameServer)
	if !success {
		logger.Error("Error registering to gameServer:", "tmpGameServer", tmpGameServer)
		return gameRtn
	}

	payload, _ := json.Marshal(game)
	requestURL = fmt.Sprintf("%s/joingame", tmpGameServer)

	body, err = m.postRequest(requestURL, payload)
	if err != nil {
		logger.Error("JoinPrivateGameRequest:", "url", requestURL, "err", err, "body", body)
		return gameRtn
	}

	gameRtn.GameId = gameId.GameId
	gameRtn.GameServer = game.Server

	return gameRtn
}

func (m Handler) HandRequest() []game.Card {
	requestURL := fmt.Sprintf("%s/hand", m.gameServer)
	var hand []game.Card

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("HandRequest:", "url", requestURL, "err", err, "body", body)
		return hand
	}

	var handStrings []string
	json.Unmarshal([]byte(body), &handStrings)

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
	requestURL := fmt.Sprintf("%s/playcard", m.gameServer)

	body, err := m.postRequest(requestURL, payload)
	if err != nil {
		logger.Error("PlayCardRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m Handler) ActionsRequest() []game.Action {
	var actions []game.Action
	requestURL := fmt.Sprintf("%s/actions", m.gameServer)

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("ActionsRequest:", "url", requestURL, "err", err, "body", body)
		return actions
	}

	json.Unmarshal([]byte(body), &actions)

	return actions
}

func (m Handler) MySeatRequest() MySeat {
	var seat MySeat
	requestURL := fmt.Sprintf("%s/seat", m.gameServer)

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("MySeatRequest:", "url", requestURL, "err", err, "body", body)
		return seat
	}

	json.Unmarshal([]byte(body), &seat)

	return seat
}

func (m Handler) ChangeTeamRequest(spectator bool) bool {
	payload := []byte("")
	if spectator {
		payload = []byte("{team:S}") // Not worth implementing the JSON.
	}

	requestURL := fmt.Sprintf("%s/changeteam", m.gameServer)

	body, err := m.postRequest(requestURL, payload)
	if err != nil {
		logger.Error("ChangeTeamRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m Handler) SwapBottomCardRequest() bool {
	requestURL := fmt.Sprintf("%s/swapBottomCard", m.gameServer)

	body, err := m.postRequest(requestURL, []byte(""))
	if err != nil {
		logger.Error("SwapBottomCardRequest:", "url", requestURL, "err", err, "body", body)
		return false
	}

	return true
}

func (m Handler) ReplayRequest(gameId GameId) []game.Action {
	requestURL := fmt.Sprintf("%s/replay?gameId=%s", m.browserServer, gameId.GameId)

	body, err := m.getRequest(requestURL)
	if err != nil {
		logger.Error("ReplayRequest:", "url", requestURL, "err", err, "body", body)
		return nil
	}

	var actions []game.Action
	json.Unmarshal([]byte(body), &actions)

	return actions
}

func (m *Handler) refreshSessionRequest() tea.Msg {
	var msg RefreshByMsg
	requestURL := fmt.Sprintf("%s/refresh", m.browserServer)

	client := &http.Client{
		Jar: m.jar,
	}

	res, err := client.Get(requestURL)
	if err != nil {
		logger.Error("refreshSessionRequest:", "url", requestURL, "err", err, "res", res)
		return nil
	}

	if res.StatusCode != http.StatusOK {
		logger.Error("refreshSessionRequest:", "url", requestURL, "err", err, "res", res)
		return nil
	}

	cookies := res.Cookies()
	for i := range cookies {
		cookie := cookies[i]
		if cookie.Name != "userId" {
			continue
		}
		msg = RefreshByMsg(cookie.Expires)
		logger.Debug("Cookies from body: ", "cookie", cookie)
	}

	client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return msg
}
