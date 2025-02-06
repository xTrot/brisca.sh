package main

import (
	"fmt"
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
	inactiveColor  = lipgloss.Color("240")
	activeColor    = lipgloss.Color("69")
	playerBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(inactiveColor)
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

func (m *gsModel) Refresh() tea.Cmd {
	return tea.Every(m.actionCache.refreshTime, func(t time.Time) tea.Msg {
		fetched := m.userGlobal.rh.actionsRequest()
		if len(fetched) != 0 {
			log.Debug("Actions fetched: ", "fetched", fetched)
		}
		m.actionCache.actions = append(m.actionCache.actions, fetched...)
		return newActionCacheMsg(m.actionCache)
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
	spinner      spinner.Model
	index        int
	boxes        [3][3]box
	hand         []card
	selectedCard int
	actionCache  actionCache
	playerSeats  []playerModel
	table        tableModel
	gameConfig   gameConfigPayload
	statusBar    statusBarModel
	userGlobal   *userGlobal
}

func newGSModel(userGlobal *userGlobal) gsModel {
	m := gsModel{
		userGlobal: userGlobal,
	}
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
	m.boxes[0][0].style = emptyBoxStyle
	m.boxes[0][1].style = playerBoxStyle
	m.boxes[0][2].style = emptyBoxStyle
	m.boxes[1][0].style = emptyBoxStyle
	m.boxes[1][1].style = tableBoxStyle
	m.boxes[1][2].style = emptyBoxStyle
	m.boxes[2][0].style = emptyBoxStyle
	m.boxes[2][1].style = playerBoxStyle
	m.boxes[2][2].style = emptyBoxStyle
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
	m.statusBar = newStatusBar(m.playerSeats)
	return m
}

func (m gsModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.spinner.Tick, m.userGlobal.LastWindowSizeReplay(), m.getMySeat(),
		m.statusBar.Init(), m.updateHand(false))
}

func (m gsModel) getMySeat() tea.Cmd {
	return func() tea.Msg {
		mySeat := m.userGlobal.rh.mySeatRequest()
		return mySeat
	}
}

func (m gsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case resizeMsg:
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				m.boxes[i][j].style = msg.boxes[i][j].style
			}
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.userGlobal.sizeMsg = &msg
		return m.updateWindow(msg)
	case newActionCacheMsg:
		m.actionCache = actionCache(msg)
		cmd = m.Refresh()
		cmds = append(cmds, cmd)
		cmd = m.actionCache.ProcessAction()
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.userGlobal.rh.leaveGameRequest()
			return m, tea.Quit
		case "q":
			m.userGlobal.rh.leaveGameRequest()
			lm := newLobby(m.userGlobal)
			return lm, lm.Init()
		case "left":
			handSize := len(m.hand)
			rawMove := m.selectedCard - 1
			m.selectedCard = (rawMove%handSize + handSize) % handSize
			return m, nil
		case "right":
			m.selectedCard = (m.selectedCard + 1) % len(m.hand)
			return m, nil
		case "enter":
			cmd = m.playCard(m.selectedCard)
			cmds = append(cmds, cmd, m.updateHand(true))
			return m, tea.Batch(cmds...)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case updateHandMsg:
		m.hand = msg.hand

	case gameConfigPayload:
		m.gameConfig = msg
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		if m.gameConfig.MaxPlayers == 3 {
			m.boxes[1][2].style = m.boxes[1][2].style.BorderStyle(lipgloss.NormalBorder()) // Adding 3rd player box
		} else if m.gameConfig.MaxPlayers == 4 {
			m.boxes[1][0].style = m.boxes[1][0].style.BorderStyle(lipgloss.NormalBorder()) // Adding 2nd player box
			m.boxes[1][2].style = m.boxes[1][2].style.BorderStyle(lipgloss.NormalBorder()) // Adding 4th player box
		}
	case gameStartedPayload:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		// Each player draws 3 cards
		m.table.deckSize -= len(msg.Seats) * 3
		m.table.cardsInPlay = []card{}
		cmd = m.processSeats(msg.Seats)
		cmds = append(cmds, cmd)
	case bottomCardSelectedPayload:
		m.table.suitCard = msg.bottomCard
	case gracePeriodEndedPayload:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case swapBottomCardPayload:
		m.table.suitCard = newBottomCard(m.table.suitCard)
	case cardDrawnPayload:
		m.table.deckSize--
		m.playerSeats[msg.Seat].handSize++
	case cardPlayedPayload:
		m.table.cardsInPlay = append(m.table.cardsInPlay, msg.card)
		m.playerSeats[msg.Seat].handSize--
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case turnWonPayload:
		slices.Reverse(m.table.cardsInPlay)
		m.playerSeats[msg.Seat].scorePile = append(m.playerSeats[msg.Seat].scorePile, m.table.cardsInPlay...)
		m.playerSeats[msg.Seat].score = m.playerSeats[msg.Seat].UpdateScore()
		m.table.cardsInPlay = []card{}
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.updateHand(false))
	case gameWonPayload:
		ws := newWinScreen(&m.gameConfig, m.playerSeats, &msg, m.userGlobal)
		return ws, ws.Init()

	case seatsMsg:
		m.playerSeats = msg
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case mySeat:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		cmd := m.Refresh()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

type seatsMsg []playerModel

func (m gsModel) processSeats(seats []seat) tea.Cmd {
	return func() tea.Msg {
		var seatsMsg seatsMsg
		for i := 0; i < len(seats); i++ {
			player := newPlayerModelFromSeat(seats[i])
			// This part only works because case mySeat: happens first then seatsMsg
			adjustedSeat := (i - m.statusBar.mySeat + m.gameConfig.MaxPlayers) % m.gameConfig.MaxPlayers
			log.Debug("gsModel:", "adjustedSeat", adjustedSeat, "i", i, "m.mySeat", m.statusBar.turn, "m.gameConfig.MaxPlayers", m.gameConfig.MaxPlayers)
			if m.gameConfig.MaxPlayers == 2 {
				player.boxX = SEAT_BASED_BOXES_2P[adjustedSeat][0]
				player.boxY = SEAT_BASED_BOXES_2P[adjustedSeat][1]
			} else if m.gameConfig.MaxPlayers == 3 {
				player.boxX = SEAT_BASED_BOXES_3P[adjustedSeat][0]
				player.boxY = SEAT_BASED_BOXES_3P[adjustedSeat][1]
			} else if m.gameConfig.MaxPlayers == 4 {
				player.boxX = SEAT_BASED_BOXES_4P[adjustedSeat][0]
				player.boxY = SEAT_BASED_BOXES_4P[adjustedSeat][1]
			}
			// This part end
			seatsMsg = append(seatsMsg, player)
		}
		return seatsMsg
	}
}

type resizeMsg struct {
	boxes [3][3]box
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
		m.boxes[0][0].style = m.boxes[0][0].style.Width(thirdWidth).Height(thirdHeight)

		// Second Player(2,3), Third Player(4)
		m.boxes[0][1].style = m.boxes[0][1].style.Width(midWidth).Height(thirdHeight)

		// Blank Top Right
		m.boxes[0][2].style = m.boxes[0][2].style.Width(thirdWidth).Height(thirdHeight)

		// Second Player(4)
		m.boxes[1][0].style = m.boxes[1][0].style.Width(thirdWidth).Height(midHeight)

		// Table
		m.boxes[1][1].style = m.boxes[1][1].style.Width(midWidth).Height(midHeight)

		// Last Player(3,4)
		m.boxes[1][2].style = m.boxes[1][2].style.Width(thirdWidth).Height(midHeight)

		// Blank Bottom Left
		m.boxes[2][0].style = m.boxes[2][0].style.Width(thirdWidth).Height(thirdHeight)

		// This player
		m.boxes[2][1].style = m.boxes[2][1].style.Width(midWidth).Height(thirdHeight)

		// Blank Bottom Right
		m.boxes[2][2].style = m.boxes[2][2].style.Width(thirdWidth).Height(thirdHeight)

		return resizeMsg{boxes: m.boxes}
	}
}

func (m gsModel) View() string {
	var s string

	m.boxes[1][1].view = m.table.View()

	for i := 0; i < m.gameConfig.MaxPlayers; i++ {
		x := m.playerSeats[i].boxX
		y := m.playerSeats[i].boxY
		m.boxes[x][y].view = m.playerSeats[i].View()
		if m.statusBar.turn == i {
			m.boxes[x][y].style = m.boxes[x][y].style.
				BorderForeground(activeColor)
		} else {
			m.boxes[x][y].style = m.boxes[x][y].style.
				BorderForeground(inactiveColor)
		}
	}

	for i := 0; i < len(m.boxes); i++ {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			m.boxes[i][0].style.Render(m.boxes[i][0].view),
			m.boxes[i][1].style.Render(m.boxes[i][1].view),
			m.boxes[i][2].style.Render(m.boxes[i][2].view),
		)
		s = lipgloss.JoinVertical(lipgloss.Top, s, row)
	}
	s = lipgloss.JoinVertical(lipgloss.Top, s, m.handView())
	s = lipgloss.JoinVertical(lipgloss.Top, s, lipgloss.JoinHorizontal(lipgloss.Left, m.statusBar.View()))
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

func (m *gsModel) updateHand(delay bool) tea.Cmd {
	return func() tea.Msg {
		if delay {
			time.Sleep(time.Millisecond * 500)
		}
		newHand := m.userGlobal.rh.handRequest()

		return updateHandMsg{
			hand: newHand,
		}
	}
}

func (m *gsModel) playCard(i int) tea.Cmd {
	if m.statusBar.isMyTurn() {
		return func() tea.Msg {
			if len(m.hand) == 1 {
				m.hand = []card{}
			} else {
				m.hand = append(m.hand[:i], m.hand[i+1:]...)
			}
			index := handIndex{Index: i}
			m.userGlobal.rh.playCardRequest(index)
			return nil
		}
	}
	return nil
}
