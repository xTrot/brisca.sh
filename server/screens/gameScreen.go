package screens

import (
	"fmt"
	"slices"
	"time"

	"brisca.sh/server/game"
	"brisca.sh/server/requests"
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
	actions     []game.Action
	refreshTime time.Duration
	processing  int
	processed   int
}

type newActionsMsg struct {
	actions  []game.Action
	gameOver bool
}

func (m *gsModel) Refresh() tea.Cmd {
	return tea.Every(m.actionCache.refreshTime, func(t time.Time) tea.Msg {
		var fetched []game.Action
		var msg newActionsMsg
		if !m.gameOver {
			fetched = m.userGlobal.ReqHandler.ActionsRequest()
		}
		msg.actions, msg.gameOver = injectClientActions(fetched)
		return msg
	})
}

func injectClientActions(fetched []game.Action) ([]game.Action, bool) {
	var effective []game.Action
	var before bool
	var clientAction game.Action
	var gameOver bool
	for _, a := range fetched {
		switch a.Payload.(type) {
		case game.CardPlayedPayload:
			before = false
			clientAction = game.Action{Type: "turn_switch", Payload: game.TurnSwitchPayload{}}
		case game.GameWonPayload:
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
			ProcessAction(m.statusBar.isMyTurn(), m.gameOver, m.statusBar.mySeat)
		return cmd
	} else {
		return nil
	}
}

type gsModel struct {
	spinner      spinner.Model
	index        int
	boxes        [3][3]box
	hand         []game.Card
	selectedCard int
	actionCache  actionCache
	playerSeats  []playerModel
	table        tableModel
	gameConfig   game.GameConfigPayload
	statusBar    statusBarModel
	userGlobal   UserGlobal
	help         gameScreenHelpModel
	cheatSheet   MarkdownModel
	showCheat    bool
	gameOver     bool
}

func newReplayGSModel(userGlobal UserGlobal, actions []game.Action) gsModel {
	m := newGSModel(userGlobal)

	m.actionCache.actions, m.gameOver = injectClientActions(actions)

	return m
}

func newGSModel(userGlobal UserGlobal) gsModel {
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
		actions:     []game.Action{},
		refreshTime: time.Millisecond * 200,
		processing:  -1,
		processed:   -1,
	}
	m.hand = []game.Card{}
	m.playerSeats = []playerModel{
		newPlayerModel(m.userGlobal.RenderEmoji),
		newPlayerModel(m.userGlobal.RenderEmoji),
		newPlayerModel(m.userGlobal.RenderEmoji),
		newPlayerModel(m.userGlobal.RenderEmoji),
	}
	m.table = newTableModel(userGlobal.RenderEmoji)
	m.statusBar = newStatusBar(m.playerSeats, userGlobal.RenderEmoji)
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
		mySeat := m.userGlobal.ReqHandler.MySeatRequest()
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
		m.userGlobal.SizeMsg = msg
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
	case requests.RefreshByMsg:
		m.userGlobal.ReqHandler.RefreshBy = time.Time(msg)
	case tea.KeyMsg:
		cmd = m.refreshSessionCheck()
		cmds = append(cmds, cmd)
		switch {
		case key.Matches(msg, m.help.keys.Quit):
			m.userGlobal.ReqHandler.LeaveGameRequest()
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
	case game.GameConfigPayload:
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
	case game.GameStartedPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		// Each player draws 3 cards
		m.table.deckSize -= len(msg.Seats) * 3
		if m.gameConfig.MaxPlayers == 3 {
			m.table.deckSize -= 1
		}
		m.table.cardsInPlay = []game.Card{}
		cmd = m.processSeats(msg.Seats)
		cmds = append(cmds, cmd)
	case game.BottomCardSelectedPayload:
		m.actionCache.processed++
		m.statusBar.swapCard = game.NewCard(msg.Card.SuitString + ":2")
		m.table.bottomCard = msg.Card
	case game.GracePeriodEndedPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case game.SwapBottomCardPayload:
		m.actionCache.processed++
		m.table.bottomCard = game.NewBottomCard(m.table.bottomCard)
		cmds = append(cmds, m.updateHand(false))
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case game.CardDrawnPayload:
		m.actionCache.processed++
		m.table.deckSize--
		m.playerSeats[msg.Seat].handSize++
		cmds = append(cmds, m.updateHand(false))
	case game.CardPlayedPayload:
		m.actionCache.processed++
		m.table.cardsInPlay = append(m.table.cardsInPlay, msg.Card)
		m.playerSeats[msg.Seat].handSize--
	case game.TurnSwitchPayload:
		m.actionCache.processed++
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case game.TurnWonPayload:
		m.actionCache.processed++
		slices.Reverse(m.table.cardsInPlay)
		m.playerSeats[msg.Seat].scorePile = append(m.playerSeats[msg.Seat].scorePile, m.table.cardsInPlay...)
		m.playerSeats[msg.Seat].score = m.playerSeats[msg.Seat].UpdateScore()
		m.table.cardsInPlay = []game.Card{}
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case game.GameWonPayload:
		m.actionCache.processed++
		m.userGlobal.LastRegisteredAction = time.Now()
		ws := newWinScreen(&m.gameConfig, m.playerSeats, &msg, m.userGlobal)
		return ws, ws.Init()
	case game.UndefinedActionPayload:
		m.actionCache.processed++
	case game.SeatAfkPayload:
		m.actionCache.processed++
		m.playerSeats[msg.Seat].afk = true
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
	case game.SeatNotAfkPayload:
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
	case requests.MySeat:
		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.Refresh())
	}

	return m, tea.Batch(cmds...)
}

func (m *gsModel) refreshSessionCheck() tea.Cmd {
	return func() tea.Msg {
		return m.userGlobal.ReqHandler.RefreshSessionCheck(14 * time.Minute)
	}
}

type seatsMsg []playerModel

func (m gsModel) processSeats(seats []game.Seat) tea.Cmd {
	return func() tea.Msg {
		var seatsMsg seatsMsg
		for i := range seats {
			player := newPlayerModelFromSeat(seats[i], m.userGlobal.RenderEmoji)
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
			s += fmt.Sprintf("%2d:%s", i+1, selectedCardStyle.Render(card.RenderCard(m.userGlobal.RenderEmoji)))
		} else {
			s += fmt.Sprintf("%2d:%s", i+1, card.RenderCard(m.userGlobal.RenderEmoji))
		}
	}
	return s
}

type updateHandMsg struct {
	hand []game.Card
}

type localUpdateHandMsg struct {
	hand []game.Card
}

func (m *gsModel) updateHand(delay bool) tea.Cmd {
	return func() tea.Msg {
		if delay {
			time.Sleep(time.Millisecond * 500)
		}
		if m.gameOver {
			return nil
		}
		newHand := m.userGlobal.ReqHandler.HandRequest()

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
			newHand := []game.Card{}
			if len(m.hand) != 1 {
				for i := range m.hand {
					if i == index {
						continue
					}
					newHand = append(newHand, (m.hand)[i])
				}
			}
			if !m.userGlobal.ReqHandler.PlayCardRequest(index) {
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
			if !m.userGlobal.ReqHandler.SwapBottomCardRequest() {
				return nil
			}
			return nil
		}
	}
	return nil
}

func (m *gsModel) swapCheck() {
	if m.table.deckSize > 1 && slices.ContainsFunc(m.hand, func(c game.Card) bool {
		return c.Num == m.statusBar.swapCard.Num && c.CharSuit == m.statusBar.swapCard.CharSuit
	}) {
		m.statusBar.canSwap = true
		m.help.keys.showSwap = true
	} else {
		m.statusBar.canSwap = false
		m.help.keys.showSwap = false
	}
}
