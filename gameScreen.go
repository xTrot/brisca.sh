package main

import (
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	emptyBoxStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	tableBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("82"))
	playerBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	activePlayerBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	gsHelpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("241"))
	selectedCardStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("69"))
	windowWidthMin  = 80
	windowHighttMin = 24
)

type box struct {
	view  string
	style lipgloss.Style
}

type actionCache struct {
	actions     []action
	refreshTime time.Duration
	processed   int
}

type newActionCacheMsg actionCache

func (ac *actionCache) Refresh() tea.Cmd {
	return tea.Every(ac.refreshTime, func(t time.Time) tea.Msg {
		fetched := actionsRequest()
		if len(fetched) != 0 {
			log.Debug("Actions fetched: ", "fetched", fetched)
		}
		ac.actions = append(ac.actions, fetched...)
		return newActionCacheMsg(*ac)
	})
}

func (ac *actionCache) ProcessAction() tea.Cmd {
	if ac.processed < len(ac.actions) {
		cmd := ac.actions[ac.processed].processAction()
		ac.processed++
		return cmd
	} else {
		return nil
	}
}

type gsModel struct {
	timer        timer.Model
	spinner      spinner.Model
	index        int
	boxes        [3][3]box
	hand         []card
	selectedCard int
	actionCache  actionCache
	playerSeats  []playerModel
	table        tableModel
	gameConfig   gameConfigPayload
	turn         int
	won          gameWonPayload
}

func newGSModel(timeout time.Duration) gsModel {
	m := gsModel{}
	m.timer = timer.New(timeout)
	m.spinner = spinner.New()
	var boxes [3][3]box
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			boxes[i][j] = box{
				view: " ",
			}
		}
	}
	m.boxes = boxes
	m.selectedCard = 0
	m.actionCache = actionCache{
		actions:     []action{},
		refreshTime: time.Millisecond * 200,
		processed:   0,
	}
	m.playerSeats = []playerModel{
		newPlayerModel(),
		newPlayerModel(),
		newPlayerModel(),
		newPlayerModel(),
	}
	m.table = newTableModel()
	return m
}

func (m gsModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.timer.Init(), m.spinner.Tick, tea.WindowSize(), m.actionCache.Refresh())
}

func (m gsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case resizeMsg:
		return msg.model, nil
	case tea.WindowSizeMsg:
		return m.updateWindow(msg)
	case newActionCacheMsg:
		m.actionCache = actionCache(msg)
		cmd = m.actionCache.Refresh()
		cmds = append(cmds, cmd)
		cmd = m.actionCache.ProcessAction()
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			handSize := len(m.hand)
			rawMove := m.selectedCard - 1
			m.selectedCard = (rawMove%handSize + handSize) % handSize
			return m, nil
		case "right":
			m.selectedCard = (m.selectedCard + 1) % len(m.hand)
			return m, nil
		case "enter":
			cmd = playCard(m.selectedCard)
			cmds = append(cmds, cmd, updateHand)
			return m, tea.Batch(cmds...)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd, updateHand)
	case updateHandMsg:
		m.hand = msg.hand

	case gameConfigPayload:
		m.gameConfig = msg
	case gameStartedPayload:
		m.turn = msg.StartingSeat
		// Each player draws 3 cards
		m.table.deckSize -= len(msg.Seats) * 3
		m.table.cardsInPlay = []card{}
		log.Debug("case gameStartedPayload:", "m.table", m.table)
		cmd = processSeats(msg.Seats)
		cmds = append(cmds, cmd)
	case bottomCardSelectedPayload:
		m.table.suitCard = msg.bottomCard
		log.Debug("case bottomCardSelectedPayload:", "m.table", m.table)
	case gracePeriodEndedPayload:
	case swapBottomCardPayload:
		m.table.suitCard = newBottomCard(m.table.suitCard)
		log.Debug("case swapBottomCardPayload:", "m.table", m.table)
	case cardDrawnPayload:
		m.table.deckSize--
		log.Debug("case cardDrawnPayload:", "m.table", m.table)
		m.playerSeats[msg.Seat].handSize++
	case cardPlayedPayload:
		m.table.cardsInPlay = append(m.table.cardsInPlay, msg.card)
		log.Debug("case cardPlayedPayload:", "m.table", m.table)
		m.playerSeats[msg.Seat].handSize--
		m.turn = (m.turn + 1) % m.gameConfig.MaxPlayers
	case turnWonPayload:
		slices.Reverse(m.table.cardsInPlay)
		m.playerSeats[msg.Seat].scorePile = append(m.playerSeats[msg.Seat].scorePile, m.table.cardsInPlay...)
		m.playerSeats[msg.Seat].score = m.playerSeats[msg.Seat].UpdateScore()
		m.table.cardsInPlay = []card{}
		m.turn = msg.Seat
	case gameWonPayload:
		m.won = msg

	case seatsMsg:
		m.playerSeats = msg
	default:
		log.Debug("default:", "msg", msg, "msg.(type)", reflect.TypeOf(msg))
	}
	return m, tea.Batch(cmds...)
}

type seatsMsg []playerModel

func processSeats(seats []seat) tea.Cmd {
	return func() tea.Msg {
		var seatsMsg seatsMsg
		for i := 0; i < len(seats); i++ {
			seatsMsg = append(seatsMsg, newPlayerModelFromSeat(seats[i]))
		}
		return seatsMsg
	}
}

type resizeMsg struct {
	model gsModel
}

func (m gsModel) updateWindow(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		wholeWidth := max(msg.Width, windowWidthMin) - 6        // 6 to account for borders
		wholeHeight := max(msg.Height, windowHighttMin) - 6 - 3 // Space reserved for bars
		thirdWidth := wholeWidth / 3
		thirdHeight := wholeHeight / 3
		midWidth := wholeWidth - (2 * thirdWidth)
		midHeight := wholeHeight - (2 * thirdHeight)

		// Blank Top Left
		m.boxes[0][0].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		// Second Player(2,3), Third Player(4)
		m.boxes[0][1].style = playerBoxStyle.Width(midWidth).Height(thirdHeight)

		// Blank Top Right
		m.boxes[0][2].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		// Second Player(4)
		m.boxes[1][0].style = emptyBoxStyle.Width(thirdWidth).Height(midHeight)

		// Table
		m.boxes[1][1].style = tableBoxStyle.Width(midWidth).Height(midHeight)

		// Last Player(3,4)
		m.boxes[1][2].style = emptyBoxStyle.Width(thirdWidth).Height(midHeight)

		// Blank Bottom Left
		m.boxes[2][0].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		// This player
		m.boxes[2][1].style = activePlayerBoxStyle.Width(midWidth).Height(thirdHeight)

		// Blank Bottom Right
		m.boxes[2][2].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		return resizeMsg{model: m}
	}
}

func (m gsModel) View() string {
	var s string
	m.boxes[0][1].view = m.playerSeats[1].View()
	// m.boxes[0][2].view =
	// m.boxes[1][0].view =
	m.boxes[1][1].view = m.table.View()
	// m.boxes[1][2].view =
	// m.boxes[2][0].view =
	m.boxes[2][1].view = m.playerSeats[0].View()
	// m.boxes[2][2].view =
	for i := 0; i < len(m.boxes); i++ {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			m.boxes[i][0].style.Render(m.boxes[i][0].view),
			m.boxes[i][1].style.Render(m.boxes[i][1].view),
			m.boxes[i][2].style.Render(m.boxes[i][2].view),
		)
		s = lipgloss.JoinVertical(lipgloss.Top, s, row)
	}
	s = lipgloss.JoinVertical(lipgloss.Top, s, m.handView())
	s = lipgloss.JoinVertical(lipgloss.Top, s, lipgloss.JoinHorizontal(lipgloss.Left, "Status: Your Turn, timer: ", m.timer.View()))
	s = lipgloss.JoinVertical(lipgloss.Center, s, gsHelpStyle.Render("← →: select card      enter: play card"))
	return s
}

func (m *gsModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *gsModel) handView() string {
	var s string
	s = "Hand:"
	for i := range len(m.hand) {
		card := m.hand[i]
		if m.selectedCard == i {
			s += fmt.Sprintf("%2d:%s", i+1, selectedCardStyle.Render(renderCard(card)))
		} else {
			s += fmt.Sprintf("%2d:%s", i+1, renderCard(card))
		}
	}
	return s
}

func renderCard(card card) string {
	return fmt.Sprintf("[%s:%2d]", card.suit, card.num)
}

type updateHandMsg struct {
	hand []card
}

func updateHand() tea.Msg {
	newHand := handRequest()

	return updateHandMsg{
		hand: newHand,
	}
}

func playCard(i int) tea.Cmd {
	index := handIndex{Index: i}
	return func() tea.Msg {
		playCardRequest(index)
		return nil
	}
}
