package screens

import (
	"strings"
	"time"

	"brisca.sh/server/game"
	"brisca.sh/server/requests"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type joinGameModel struct {
	form      *huh.Form // huh.Form is just a tea.Model
	nextView  tea.Model
	usc       UserScreenContext
	gameId    *string
	replay    bool
	waiting   bool
	waitStyle lipgloss.Style
	spinner   spinner.Model
}

func newReplayGame(nv tea.Model, usc UserScreenContext) joinGameModel {
	var gameId string
	return joinGameModel{
		gameId: &gameId,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What's the UUID of the Game you want to replay?").
					Value(&gameId),
			),
		),
		nextView: nv,
		usc:      usc,
		replay:   true,
		waiting:  false,
		waitStyle: usc.Renderer().NewStyle().
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(
				usc.Renderer().NewStyle().
					AlignHorizontal(lipgloss.Center).
					AlignVertical(lipgloss.Center),
			),
		),
	}
}

func newJoinGame(nv tea.Model, usc UserScreenContext) joinGameModel {
	var gameId string
	return joinGameModel{
		gameId: &gameId,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What's the UUID of the Game you want to join?").
					Value(&gameId),
			),
		),
		nextView: nv,
		usc:      usc,
		replay:   false,
		waiting:  false,
		waitStyle: usc.Renderer().NewStyle().
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(
				usc.Renderer().NewStyle().
					AlignHorizontal(lipgloss.Center).
					AlignVertical(lipgloss.Center),
			),
		),
	}
}

func (m joinGameModel) Init() tea.Cmd {
	return tea.Batch(
		m.form.Init(),
		m.spinner.Tick,
		m.usc.LastWindowSizeReplay(),
	)
}

type replayMsg []game.Action
type joinedMsg requests.NewGame

func (m joinGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ...

	var cmds []tea.Cmd
	quit, cmd := HasRetiredUser(&m.usc)
	if cmd != nil {
		return quit, cmd
	}

	if time.Now().After(m.usc.ReqHandler().RefreshBy) {
		log.Debug("Idle Disconnect")
		return NewModel("Idle Disconnect")
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	cmds = append(cmds, cmd)

	if m.form.State == huh.StateCompleted && !m.waiting {
		var gameId requests.GameId
		gameId.GameId = strings.TrimSpace(*m.gameId)

		if m.replay {

			cmd = func() tea.Msg {
				log.Debug("Requesting replay:",
					"m.usc.Username", m.usc.Username,
					"gameId", gameId,
				)
				time.Sleep(time.Second)
				rtn := replayMsg(
					m.usc.ReqHandler().ReplayRequest(gameId),
				)
				log.Debug("Result:", "rtn", rtn)
				return rtn
			}

		} else {

			cmd = func() tea.Msg {
				log.Debug("Requesting joinPrivateGame:",
					"m.usc.Username", m.usc.Username,
					"gameId", gameId,
				)
				time.Sleep(time.Second)
				rtn := joinedMsg(
					m.usc.ReqHandler().JoinPrivateGameRequest(
						gameId,
						m.usc.Username(),
					),
				)
				log.Debug("Result:", "rtn", rtn)
				return rtn
			}

		}

		cmds = append(cmds, cmd)
		m.waiting = true

	}

	switch msg := msg.(type) {

	case replayMsg:
		if msg != nil {
			rgs := newReplayGSModel(m.usc, msg)
			return rgs, rgs.Init()
		} else {
			return m.nextView, m.nextView.Init()
		}

	case joinedMsg:
		emptyGame := requests.NewGame{}
		newGame := requests.NewGame(msg)
		if emptyGame != newGame {
			m.usc.ReqHandler().SetGameServer(newGame.GameServer)
			wrm := newWaitingRoom(m.usc)
			wrm.list.Title = "GameID: " + *m.gameId
			cmd = wrm.Init()
			return wrm, cmd
		} else {
			return m.nextView, m.nextView.Init()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.usc.SetWindowsSize(msg)
		m.waitStyle = m.waitStyle.
			Height(msg.Height).
			Width(msg.Width)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)

}

func (m joinGameModel) View() string {
	if m.waiting {

		var waitingFor string
		if m.replay {
			waitingFor = " Fetching replay"
		} else {
			waitingFor = " Joining game"
		}

		return m.waitStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.spinner.View(),
				waitingFor,
			),
		)

	} else {
		return m.form.View()
	}

}
