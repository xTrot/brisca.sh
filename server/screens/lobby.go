package screens

import (
	"time"

	"brisca.sh/server/requests"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	lobbyIsStale = 5 //seconds
	testDelay    = 1 //seconds
)

var ()

type listKeyMap struct {
	insertItem key.Binding
	joinGame   key.Binding
	replayGame key.Binding
	choose     key.Binding
	help       key.Binding
	emoji      key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		joinGame: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "private game(join)"),
		),
		replayGame: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "replay game"),
		),
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		help: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "how to play"),
		),
		emoji: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "toggle emoji rendering"),
		),
	}
}

type lobbyModel struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	lastUpdate   time.Time
	usc          UserScreenContext
	fullHelp     MarkdownModel
	showFH       bool

	docStyle           lipgloss.Style
	statusMessageStyle lipgloss.Style
}

type itemsMsg struct {
	items []list.Item
}

func NewLobby(usc UserScreenContext) lobbyModel {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Make initial list of items
	const numItems = 0
	items := make([]list.Item, numItems)

	// Setup list

	lm := lobbyModel{}

	delegate := newItemDelegate(delegateKeys, &lm)
	gamesList := list.New(items, delegate, 0, 0)
	gamesList.Styles.Title = usc.Renderer().NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
	gamesList.Title = "brisca.sh games:"
	gamesList.SetStatusBarItemName("game", "games")
	gamesList.Help = help.New()
	gamesList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.insertItem,
			listKeys.joinGame,
			listKeys.replayGame,
			listKeys.choose,
			listKeys.help,
			listKeys.emoji,
		}
	}
	gamesList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.insertItem,
			listKeys.joinGame,
			listKeys.choose,
			listKeys.help,
		}
	}
	gamesList.KeyMap.Filter.SetHelp("/", "search")

	lm.list = gamesList
	lm.keys = listKeys
	lm.delegateKeys = delegateKeys
	lm.lastUpdate = time.Now()
	lm.usc = usc
	lm.fullHelp = NewFullHelpModel()

	lm.docStyle = lm.usc.Renderer().NewStyle().Margin(1, 2)
	lm.statusMessageStyle = lm.usc.Renderer().NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})
	return lm
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (lm lobbyModel) Init() tea.Cmd {
	_, cmd := lm.updateIfStale(0)
	return tea.Batch(cmd, doTick(), lm.usc.LastWindowSizeReplay(), lm.list.StartSpinner())
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	quit, cmd := HasRetiredUser(&m.usc)
	if cmd != nil {
		return quit, cmd
	}

	if time.Now().After(m.usc.ReqHandler().RefreshBy) {
		logger.Debug("Idle Disconnect")
		return NewModel("Idle Disconnect")
	}

	switch msg := msg.(type) {

	case joinGameMsg:
		wrm := newWaitingRoom(msg.usc)
		wrm.list.Title = "GameID: " + msg.gameId.GameId
		cmd = wrm.Init()
		return wrm, cmd

	case tickMsg:
		m, cmd = m.updateIfStale(lobbyIsStale)
		cmds = append(cmds, cmd, doTick())

	case itemsMsg:
		cmd = m.list.SetItems(msg.items)
		m.lastUpdate = time.Now()
		cmds = append(cmds, cmd)
		m.list.StopSpinner()
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		m.usc.SetWindowsSize(msg)
		h, v := m.docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		var model tea.Model
		model, cmd = m.fullHelp.Update(msg)
		if fullHelp, ok := model.(MarkdownModel); ok {
			m.fullHelp = fullHelp
		}
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		cmd = m.refreshSessionCheck()
		cmds = append(cmds, cmd)
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.help):
			m.showFH = !m.showFH
		default:
			if m.showFH {
				var model tea.Model
				model, cmd = m.fullHelp.Update(msg)
				if fullHelp, ok := model.(MarkdownModel); ok {
					m.fullHelp = fullHelp
				}
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		if m.showFH {
			break
		}

		switch {
		case key.Matches(msg, m.keys.insertItem):
			mg := newMakeGame(m, m.usc)
			return mg, mg.Init()
		case key.Matches(msg, m.keys.joinGame):
			jg := newJoinGame(m, m.usc)
			return jg, jg.Init()
		case key.Matches(msg, m.keys.replayGame):
			rg := newReplayGame(m, m.usc)
			return rg, rg.Init()
		case key.Matches(msg, m.keys.emoji):
			if m.usc.RenderEmoji() {
				m.list.StatusMessageLifetime = time.Second * 2
				m.usc.SetRenderEmoji(false)
				cmds = append(cmds, m.list.NewStatusMessage("Emoji rendering disabled."))
			} else {
				m.list.StatusMessageLifetime = time.Second * 2
				m.usc.SetRenderEmoji(true)
				cmds = append(cmds, m.list.NewStatusMessage("Emoji rendering enabled."))
			}
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (lm lobbyModel) refreshSessionCheck() tea.Cmd {
	return func() tea.Msg {
		return lm.usc.ReqHandler().RefreshSessionCheck(5 * time.Minute)
	}
}

func (lm lobbyModel) View() string {
	if lm.showFH {
		return lm.fullHelp.View()
	}
	return lm.docStyle.Render(lm.list.View())
}

func (lm lobbyModel) updateIfStale(stale int) (lobbyModel, tea.Cmd) {
	if lm.list.FilterState() == list.Filtering {
		return lm, nil // Don't update midsearch
	}

	currentTime := time.Now()

	diff := currentTime.Sub(lm.lastUpdate)
	if diff >= (time.Second * time.Duration(stale)) {
		cmd := lm.list.StartSpinner()

		return lm, tea.Batch(cmd, func() tea.Msg {
			time.Sleep(time.Second * testDelay)
			newItems := lm.usc.ReqHandler().LobbyRequest()
			return itemsMsg{
				items: newItems,
			}
		})
	}

	return lm, nil
}

type joinGameMsg struct {
	gameId requests.GameId
	usc    UserScreenContext
}

func (m *lobbyModel) joinGame(game requests.Game) tea.Cmd {
	return func() tea.Msg {
		gameId := requests.GameId{GameId: game.GameId}
		emptyGame := requests.NewGame{}
		newGame := m.usc.ReqHandler().JoinGameRequest(gameId, game.Server, m.usc.Username())
		if emptyGame != newGame {
			m.usc.ReqHandler().SetGameServer(newGame.GameServer)
			return joinGameMsg{
				gameId: gameId,
				usc:    m.usc,
			}
		} else {
			return nil
		}
	}
}
