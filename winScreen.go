package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	winScreenStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69"))
	winnerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder()).
			BorderForeground(lipgloss.Color("69"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(75).Height(5).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
)

type winScreen struct {
	scArray    [3]scoreCounter
	scSize     int
	countDone  int
	style      lipgloss.Style
	gameConfig gameConfigPayload
	winString  string
	userGlobal *userGlobal
}

type doneCounting struct {
	index int
}

func newWinScreen(gc *gameConfigPayload, players []playerModel, gameWon *gameWonPayload, userGlobal *userGlobal) winScreen {

	var firstScoreCounter scoreCounter
	var secondScoreCounter scoreCounter
	var thirdScoreCounter scoreCounter
	var scSize int
	var winString string

	switch gc.MaxPlayers {
	case 2:
		firstScoreCounter = newScoreCounter(0, players[0].name, players[0].scorePile)
		secondScoreCounter = newScoreCounter(1, players[1].name, players[1].scorePile)
		scSize = 2
		winString = players[gameWon.Seat].name + " won!!!"
	case 3:
		firstScoreCounter = newScoreCounter(0, players[0].name, players[0].scorePile)
		secondScoreCounter = newScoreCounter(1, players[1].name, players[1].scorePile)
		thirdScoreCounter = newScoreCounter(2, players[2].name, players[2].scorePile)
		scSize = 3
		winString = players[gameWon.Seat].name + " won!!!"
	case 4:
		firstScoreCounter = newScoreCounter(0, "Team A", append(players[0].scorePile, players[2].scorePile...))
		secondScoreCounter = newScoreCounter(1, "Team B", append(players[1].scorePile, players[3].scorePile...))
		scSize = 2
		winString = "Team " + gameWon.Team + " won!!!"
	default:
		panic(fmt.Sprintf("gameConfig.MaxPlayers not 2-4, gameConfig=%v", gc))
	}

	return winScreen{
		style:      winScreenStyle,
		gameConfig: *gc,
		scArray: [3]scoreCounter{
			firstScoreCounter,
			secondScoreCounter,
			thirdScoreCounter,
		},
		scSize:     scSize,
		winString:  winString,
		userGlobal: userGlobal,
	}
}

func (m winScreen) Init() tea.Cmd {
	if m.gameConfig.MaxPlayers == 3 {
		return tea.Batch(
			m.userGlobal.LastWindowSizeReplay(),
			m.scArray[0].Init(),
			m.scArray[1].Init(),
			m.scArray[2].Init(),
		)
	} else {
		return tea.Batch(
			m.userGlobal.LastWindowSizeReplay(),
			m.scArray[0].Init(),
			m.scArray[1].Init(),
		)
	}
}

func (m winScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.userGlobal.sizeMsg = &msg
		m.style = m.style.Width(msg.Width - 2).Height(msg.Height - 2)

	case tea.KeyMsg:
		lm := newLobby(m.userGlobal)
		return lm, lm.Init()
	case pretendCountMsg:
		switch msg.id {
		case 0:
			m.scArray[0], cmd = m.scArray[0].Update(msg)
			cmds = append(cmds, cmd)
		case 1:
			m.scArray[1], cmd = m.scArray[1].Update(msg)
			cmds = append(cmds, cmd)
		case 2:
			m.scArray[2], cmd = m.scArray[2].Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.scArray[0], cmd = m.scArray[0].Update(msg)
		cmds = append(cmds, cmd)
		m.scArray[1], cmd = m.scArray[1].Update(msg)
		cmds = append(cmds, cmd)
		m.scArray[2], cmd = m.scArray[2].Update(msg)
		cmds = append(cmds, cmd)
	case doneCounting:
		m.countDone++
		log.Debug("case doneCounting:", "msg", msg)
	}

	return m, tea.Batch(cmds...)
}

func (m winScreen) View() string {
	var s string
	if m.gameConfig.MaxPlayers == 3 {
		s = lipgloss.JoinHorizontal(lipgloss.Center,
			m.scArray[0].View(), m.scArray[1].View(), m.scArray[2].View())
	} else {
		s = lipgloss.JoinHorizontal(lipgloss.Center,
			m.scArray[0].View(), m.scArray[1].View())
	}

	winner := m.winString
	if m.countDone != m.scSize {
		winner = " "
	}

	s = lipgloss.JoinVertical(lipgloss.Center,
		winnerStyle.Render("\n"+winner+"\n"),
		s,
		helpStyle.AlignHorizontal(lipgloss.Center).Render("Press any key to exit"))

	return m.style.Render(s)
}
