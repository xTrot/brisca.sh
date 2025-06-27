package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
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
	processing  int
	processed   int
}

type newActionsMsg struct {
	actions  []action
	gameOver bool
}

func (m *gsModel) Refresh() tea.Cmd {
	return tea.Every(m.actionCache.refreshTime, func(t time.Time) tea.Msg {
		var fetched []action
		var msg newActionsMsg
		if !m.gameOver {
			fetched = m.userGlobal.rh.actionsRequest()
		}
		msg.actions, msg.gameOver = injectClientActions(fetched)
		return msg
	})
}

func injectClientActions(fetched []action) ([]action, bool) {
	var effective []action
	var before bool
	var clientAction action
	var gameOver bool
	for _, a := range fetched {
		switch a.Payload.(type) {
		case cardPlayedPayload:
			before = false
			clientAction = action{Type: "turn_switch", Payload: turnSwitchPayload{}}
		case gameWonPayload:
			gameOver = true
		default:
			effective = append(effective, a)
			continue
		}
		if before {
			effective = append(effective, clientAction)
			effective = append(effective, a)
		} else {
			effective = append(effective, a)
			effective = append(effective, clientAction)
		}
	}
	return effective, gameOver
}

func (m *gsModel) ProcessAction() tea.Cmd {
	if len(m.actionCache.actions) > m.actionCache.processing+1 &&
		m.actionCache.processing == m.actionCache.processed {
		m.actionCache.processing++
		cmd := m.actionCache.actions[m.actionCache.processing].
			processAction(m.statusBar.isMyTurn(), m.gameOver, m.statusBar.mySeat)
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
	userGlobal   userGlobal
	help         gameScreenHelpModel
	cheatSheet   MarkdownModel
	showCheat    bool
	gameOver     bool
}

func newReplayGSModel(userGlobal userGlobal, actions []action) gsModel {
	m := newGSModel(userGlobal)

	m.actionCache.actions, m.gameOver = injectClientActions(actions)

	return m
}

func newGSModel(userGlobal userGlobal) gsModel {
	m := gsModel{
		userGlobal: userGlobal,
	}
	m.spinner = spinner.New()
	var boxes [3][3]box
	for i := range 3 {
		for j := range 3 {
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
		processing:  -1,
		processed:   -1,
	}
	m.hand = []card{}
	m.playerSeats = []playerModel{
		newPlayerModel(m.userGlobal.renderEmoji),
		newPlayerModel(m.userGlobal.renderEmoji),
		newPlayerModel(m.userGlobal.renderEmoji),
		newPlayerModel(m.userGlobal.renderEmoji),
	}
	m.table = newTableModel(userGlobal.renderEmoji)
	m.statusBar = newStatusBar(m.playerSeats, userGlobal.renderEmoji)
	m.help = newGSHelp()
	m.cheatSheet = NewCheatSheetModel()
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
		for i := range 3 {
			for j := range 3 {
				m.boxes[i][j].style = msg.boxes[i][j].style
			}
		}
		m.cheatSheet.Style = msg.csStyle
		return m, nil
	case tea.WindowSizeMsg:
		m.userGlobal.sizeMsg = msg
		return m.updateWindow(msg)
	case newActionsMsg:
		if len(msg.actions) > 0 {
			m.actionCache.actions = append(m.actionCache.actions, msg.actions...)
		}
		if !m.gameOver && msg.gameOver {
			m.gameOver = msg.gameOver
			log.Debug("gameOver:", "gameId", m.gameConfig.GameId)
		}
		cmd = m.Refresh()
		cmds = append(cmds, cmd)
		cmd = m.ProcessAction()
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		cmd = m.refreshSessionCheck()
		cmds = append(cmds, cmd)
		switch {
		case key.Matches(msg, m.help.keys.Quit):
			m.userGlobal.rh.leaveGameRequest()
			return m, tea.Quit
		// case "q":
		// 	m.userGlobal.rh.leaveGameRequest()
		// 	lm := newLobby(m.userGlobal)
		// 	return lm, lm.Init()
		case key.Matches(msg, m.help.keys.Left):
			handSize := len(m.hand)
			rawMove := m.selectedCard - 1
			m.selectedCard = (rawMove%handSize + handSize) % handSize
		case key.Matches(msg, m.help.keys.Right):
			m.selectedCard = (m.selectedCard + 1) % len(m.hand)
		case key.Matches(msg, m.help.keys.Enter):
			cmd = m.playCard(m.selectedCard)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.help.keys.One):
			cmd = m.playCard(0)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.help.keys.Two):
			cmd = m.playCard(1)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.help.keys.Three):
			cmd = m.playCard(2)
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.help.keys.Swap):
			cmd = m.swapBottomCard()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.help.keys.Cheat):
			m.showCheat = !m.showCheat

			// case key.Matches(msg, m.help.keys.Help):
			// 	m.help.help.ShowAll = !m.help.help.ShowAll

		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case updateHandMsg:
		m.hand = msg.hand
		m.swapCheck()
	case localUpdateHandMsg:
		m.statusBar.iPlayed = true
		m.hand = msg.hand
		m.swapCheck()

		// All Payload case statement must update ac processed
	case gameConfigPayload:
		m.actionCache.processed++
		m.gameConfig = msg
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		switch m.gameConfig.MaxPlayers {
		case 3:
			m.boxes[1][2].style = m.boxes[1][2].style.BorderStyle(lipgloss.NormalBorder()) // Adding 3rd player box
		case 4:
			m.boxes[1][0].style = m.boxes[1][0].style.BorderStyle(lipgloss.NormalBorder()) // Adding 2nd player box
			m.boxes[1][2].style = m.boxes[1][2].style.BorderStyle(lipgloss.NormalBorder()) // Adding 4th player box
		}
	case gameStartedPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		// Each player draws 3 cards
		m.table.deckSize -= len(msg.Seats) * 3
		if m.gameConfig.MaxPlayers == 3 {
			m.table.deckSize -= 1
		}
		m.table.cardsInPlay = []card{}
		cmd = m.processSeats(msg.Seats)
		cmds = append(cmds, cmd)
	case bottomCardSelectedPayload:
		m.actionCache.processed++
		m.statusBar.swapCard = newCard(msg.bottomCard.suitString + ":2")
		m.table.bottomCard = msg.bottomCard
	case gracePeriodEndedPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case swapBottomCardPayload:
		m.actionCache.processed++
		m.table.bottomCard = newBottomCard(m.table.bottomCard)
		cmds = append(cmds, m.updateHand(false))
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case cardDrawnPayload:
		m.actionCache.processed++
		m.table.deckSize--
		m.playerSeats[msg.Seat].handSize++
		cmds = append(cmds, m.updateHand(false))
	case cardPlayedPayload:
		m.actionCache.processed++
		m.table.cardsInPlay = append(m.table.cardsInPlay, msg.card)
		m.playerSeats[msg.Seat].handSize--
	case turnSwitchPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case turnWonPayload:
		m.actionCache.processed++
		slices.Reverse(m.table.cardsInPlay)
		m.playerSeats[msg.Seat].scorePile = append(m.playerSeats[msg.Seat].scorePile, m.table.cardsInPlay...)
		m.playerSeats[msg.Seat].score = m.playerSeats[msg.Seat].UpdateScore()
		m.table.cardsInPlay = []card{}
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case gameWonPayload:
		m.actionCache.processed++
		ws := newWinScreen(&m.gameConfig, m.playerSeats, &msg, m.userGlobal)
		return ws, ws.Init()
	case undefinedActionPayload:
		m.actionCache.processed++
	case seatAfkPayload:
		m.actionCache.processed++
		m.playerSeats[msg.Seat].afk = true
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case seatNotAfkPayload:
		m.actionCache.processed++
		m.playerSeats[msg.Seat].afk = true
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)

	case seatsMsg:
		m.playerSeats = msg
		for i := range msg {
			m.boxes[msg[i].boxX][msg[i].boxY].style = playerBoxStyle
		}
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd, m.userGlobal.LastWindowSizeReplay())
	case mySeat:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.Refresh())
	}

	return m, tea.Batch(cmds...)
}

func (m *gsModel) refreshSessionCheck() tea.Cmd {
	return func() tea.Msg {
		return m.userGlobal.rh.refreshSessionCheck(10 * time.Minute)
	}
}

type seatsMsg []playerModel

func (m gsModel) processSeats(seats []seat) tea.Cmd {
	return func() tea.Msg {
		var seatsMsg seatsMsg
		for i := range seats {
			player := newPlayerModelFromSeat(seats[i], m.userGlobal.renderEmoji)
			// This part only works because case mySeat: happens first then seatsMsg
			adjustedSeat := (i - m.statusBar.mySeat + m.gameConfig.MaxPlayers) % m.gameConfig.MaxPlayers
			log.Debug("gsModel:", "adjustedSeat", adjustedSeat, "i", i, "m.mySeat", m.statusBar.turn, "m.gameConfig.MaxPlayers", m.gameConfig.MaxPlayers)
			switch m.gameConfig.MaxPlayers {
			case 2:
				player.boxX = SEAT_BASED_BOXES_2P[adjustedSeat][0]
				player.boxY = SEAT_BASED_BOXES_2P[adjustedSeat][1]
			case 3:
				player.boxX = SEAT_BASED_BOXES_3P[adjustedSeat][0]
				player.boxY = SEAT_BASED_BOXES_3P[adjustedSeat][1]
			case 4:
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
	boxes   [3][3]box
	csStyle lipgloss.Style
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

		wholeWidth = max(msg.Width, windowWidthMin) - 2    // 2 borders, 2?
		wholeHeight = max(msg.Height, windowHighttMin) - 3 // 2 borders, 1 help bar
		m.cheatSheet.Style = m.cheatSheet.Style.Width(wholeWidth).Height(wholeHeight)

		return resizeMsg{
			boxes:   m.boxes,
			csStyle: m.cheatSheet.Style,
		}
	}
}

func (m gsModel) View() string {
	var s string

	if m.showCheat {
		s = lipgloss.JoinVertical(lipgloss.Top, s,
			m.cheatSheet.Style.Render(m.cheatSheet.View()),
		)
	} else {
		m.boxes[1][1].view = m.table.View(
			m.boxes[1][1].style.GetWidth(),
			m.boxes[1][1].style.GetHeight(),
		)

		for i := range m.gameConfig.MaxPlayers {
			x := m.playerSeats[i].boxX
			y := m.playerSeats[i].boxY
			m.boxes[x][y].view = m.playerSeats[i].View(
				m.boxes[x][y].style.GetWidth(), m.boxes[x][y].style.GetHeight(),
			)
			if m.statusBar.turn == i {
				m.boxes[x][y].style = m.boxes[x][y].style.
					BorderForeground(activeColor)
			} else {
				m.boxes[x][y].style = m.boxes[x][y].style.
					BorderForeground(inactiveColor)
			}
		}

		for i := range len(m.boxes) {
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				m.boxes[i][0].style.Render(m.boxes[i][0].view),
				m.boxes[i][1].style.Render(m.boxes[i][1].view),
				m.boxes[i][2].style.Render(m.boxes[i][2].view),
			)
			s = lipgloss.JoinVertical(lipgloss.Top, s, row)
		}
		s = lipgloss.JoinVertical(lipgloss.Top, s, m.handView())
		s = lipgloss.JoinVertical(lipgloss.Top, s, lipgloss.JoinHorizontal(lipgloss.Left, m.statusBar.View(m.hand)))
	}
	s = lipgloss.JoinVertical(lipgloss.Center, s, gsHelpStyle.Render(m.help.View()))
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
	for i := range m.hand {
		card := (m.hand)[i]
		if m.selectedCard == i {
			s += fmt.Sprintf("%2d:%s", i+1, selectedCardStyle.Render(card.renderCard(m.userGlobal.renderEmoji)))
		} else {
			s += fmt.Sprintf("%2d:%s", i+1, card.renderCard(m.userGlobal.renderEmoji))
		}
	}
	return s
}

type updateHandMsg struct {
	hand []card
}

type localUpdateHandMsg struct {
	hand []card
}

func (m *gsModel) updateHand(delay bool) tea.Cmd {
	return func() tea.Msg {
		if delay {
			time.Sleep(time.Millisecond * 500)
		}
		if m.gameOver {
			return nil
		}
		newHand := m.userGlobal.rh.handRequest()

		return updateHandMsg{newHand}
	}
}

func (m *gsModel) playCard(index int) tea.Cmd {
	if m.statusBar.isMyTurn() && m.statusBar.haventPlayed() {
		return func() tea.Msg {
			handSize := len(m.hand)
			if handSize <= index {
				return nil
			}
			newHand := []card{}
			if len(m.hand) != 1 {
				for i := range m.hand {
					if i == index {
						continue
					}
					newHand = append(newHand, (m.hand)[i])
				}
			}
			index := handIndex{Index: index}
			if !m.userGlobal.rh.playCardRequest(index) {
				return nil
			}
			return localUpdateHandMsg{newHand}
		}
	}
	return nil
}

func (m *gsModel) swapBottomCard() tea.Cmd {
	if m.statusBar.isMyTurn() && m.statusBar.canSwap {
		return func() tea.Msg {
			if !m.userGlobal.rh.swapBottomCardRequest() {
				return nil
			}
			return nil
		}
	}
	return nil
}

func (m *gsModel) swapCheck() {
	if m.table.deckSize > 1 && slices.ContainsFunc(m.hand, func(c card) bool {
		return c.num == m.statusBar.swapCard.num && c.charSuit == m.statusBar.swapCard.charSuit
	}) {
		m.statusBar.canSwap = true
		m.help.keys.showSwap = true
	} else {
		m.statusBar.canSwap = false
		m.help.keys.showSwap = false
	}
}
