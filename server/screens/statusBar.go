package screens

import (
	"fmt"
	"time"

	"brisca.sh/server/game"
	"brisca.sh/server/requests"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	TURN_LENGTH     = time.Second * 59 // Actually one minute
	GRACE_LENGTH    = time.Second * 9  // Actually ten seconds
	AFK_TURN_LENGTH = time.Second * 4  // Actually five seconds
)

type statusBarModel struct {
	timer       timer.Model
	players     []playerModel
	turn        int
	hasStarted  bool
	cardsPlayed int
	iPlayed     bool
	canSwap     bool

	// config
	mySeat         int
	swapBottomCard bool
	maxPlayers     int
	swapCard       game.Card
	renderEmoji    bool
}

func (m statusBarModel) haventPlayed() bool {
	return !m.iPlayed
}

func newStatusBar(players []playerModel, renderEmoji bool) statusBarModel {
	return statusBarModel{
		timer:       timer.New(GRACE_LENGTH),
		players:     players,
		renderEmoji: renderEmoji,
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
	case requests.MySeat:
		m.mySeat = msg.Seat
	case seatsMsg:
		m.players = msg
	case game.GameStartedPayload:
		m.turn = msg.StartingSeat
		m.timer = timer.New(GRACE_LENGTH)
		cmds = append(cmds, m.timer.Init())
	case game.GracePeriodEndedPayload:
		m.hasStarted = true
		m.timer = timer.New(TURN_LENGTH)
		cmds = append(cmds, m.timer.Init())
	case game.GameConfigPayload:
		m.maxPlayers = msg.MaxPlayers
		m.swapBottomCard = msg.SwapBottomCard
	case game.TurnSwitchPayload:
		if m.maxPlayers == 0 {
			errMsg := "m.maxPlayers must be set before " +
				"case cardPlayedPayload: in statusBarModel.Update"
			logger.Error(errMsg)
			panic(errMsg)
		}
		m.cardsPlayed++
		if m.cardsPlayed < m.maxPlayers {
			m.turn = (m.turn + 1) % m.maxPlayers
		}
		var timerLength time.Duration
		if m.players[m.turn].afk {
			timerLength = AFK_TURN_LENGTH
		} else {
			timerLength = TURN_LENGTH
		}
		m.timer = timer.New(timerLength)
		cmds = append(cmds, m.timer.Init())
	case game.TurnWonPayload:
		m.iPlayed = false
		m.cardsPlayed = 0
		m.turn = msg.Seat
		var timerLength time.Duration
		if m.players[m.turn].afk {
			timerLength = AFK_TURN_LENGTH
		} else {
			timerLength = TURN_LENGTH
		}
		m.timer = timer.New(timerLength)
		cmds = append(cmds, m.timer.Init())
	case game.SeatAfkPayload:
		m.players[msg.Seat].afk = true
	case game.SeatNotAfkPayload:
		m.players[msg.Seat].afk = false
	}

	return m, tea.Batch(cmds...)
}

func (m *statusBarModel) View(hand []game.Card) string {
	if m.hasStarted {
		var turnString string
		if m.mySeat == m.turn {
			turnString = "Your Turn"
		} else {
			turnString = m.players[m.turn].name + "'s Turn"
		}

		swapCardStatus := ""
		if m.swapBottomCard && m.turn == m.mySeat && m.canSwap {
			swapCardStatus = ", you can swap " + m.swapCard.RenderCard(m.renderEmoji) + " for the life card"
		}

		return fmt.Sprintf("Status: %s, timer: %s%s", turnString, m.timer.View(), swapCardStatus)
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
