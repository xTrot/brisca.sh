package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

var (
	statusString = "Status: Your Turn, timer: %s"
	TURN_LENGTH  = time.Second * 59 // Actually one minute
	GRACE_LENGTH = time.Second * 9  // Actually ten seconds
)

type statusBarModel struct {
	timer      timer.Model
	players    []playerModel
	turn       int
	mySeat     int
	maxPlayers int
	hasStarted bool
}

func newStatusBar(players []playerModel) statusBarModel {
	return statusBarModel{
		timer:   timer.New(GRACE_LENGTH),
		players: players,
	}
}

func (m statusBarModel) Init() tea.Cmd {
	return tea.Batch(m.timer.Init())
}

func (m statusBarModel) Update(msg tea.Msg) (statusBarModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
	case mySeat:
		m.mySeat = msg.Seat
	case seatsMsg:
		m.players = msg
	case gameStartedPayload:
		m.turn = msg.StartingSeat
		m.timer = timer.New(GRACE_LENGTH)
		cmds = append(cmds, m.timer.Init())
	case gracePeriodEndedPayload:
		m.hasStarted = true
		m.timer = timer.New(TURN_LENGTH)
		cmds = append(cmds, m.timer.Init())
	case gameConfigPayload:
		m.maxPlayers = msg.MaxPlayers
	case cardPlayedPayload:
		if m.maxPlayers == 0 {
			errMsg := "m.maxPlayers must be set before " +
				"case cardPlayedPayload: in statusBarModel.Update"
			log.Fatal(errMsg)
			panic(errMsg)
		}
		m.turn = (m.turn + 1) % m.maxPlayers
		m.timer = timer.New(TURN_LENGTH)
		cmds = append(cmds, m.timer.Init())
	case turnWonPayload:
		m.turn = msg.Seat
		m.timer = timer.New(TURN_LENGTH)
		cmds = append(cmds, m.timer.Init())
	}

	return m, tea.Batch(cmds...)
}

func (m statusBarModel) View() string {
	if m.hasStarted {
		var turnString string
		if m.mySeat == m.turn {
			turnString = "Your Turn"
		} else {
			turnString = m.players[m.turn].name + "'s Turn"
		}
		return fmt.Sprintf("Status: %s, timer: %s", turnString, m.timer.View())
	} else {
		return fmt.Sprintf("Status: Grace period, timer: %s", m.timer.View())
	}
}

func (m statusBarModel) isMyTurn() bool {
	if m.hasStarted && (m.turn == m.mySeat) {
		return true
	}
	return false
}
