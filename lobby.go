package main

import (
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/compat"
)

const (
	lobbyIsStale = 5 //seconds
	testDelay    = 1 //seconds
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(compat.AdaptiveColor{Light: lipgloss.Color("#04B575"), Dark: lipgloss.Color("#04B575")}).
				Render
)

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
			key.WithKeys("j"),
			key.WithHelp("j", "join game"),
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
	userGlobal   userGlobal
	fullHelp     MarkdownModel
	showFH       bool
}

type itemsMsg struct {
	items []list.Item
}

func newLobby(userGlobal userGlobal) lobbyModel {
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
	gamesList.Styles.Title = titleStyle
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
	lm.userGlobal = userGlobal
	lm.fullHelp = NewFullHelpModel()

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
	return tea.Batch(cmd, doTick(), lm.userGlobal.LastWindowSizeReplay(), lm.list.StartSpinner())
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case joinGameMsg:
		wrm := newWaitingRoom(m.userGlobal)
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
		m.userGlobal.sizeMsg = msg
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		var model tea.Model
		model, cmd = m.fullHelp.Update(msg)
		if fullHelp, ok := model.(MarkdownModel); ok {
			m.fullHelp = fullHelp
		}
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
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
			mg := newV2MakeGame(v1LobbyModel{m}, m.userGlobal)
			return mg, mg.Init()
		case key.Matches(msg, m.keys.joinGame):
			jg := newV2JoinGame(v1LobbyModel{m}, m.userGlobal)
			return jg, jg.Init()
		case key.Matches(msg, m.keys.replayGame):
			rg := newV2ReplayGame(v1LobbyModel{m}, m.userGlobal)
			return rg, rg.Init()
		case key.Matches(msg, m.keys.emoji):
			if m.userGlobal.renderEmoji {
				m.list.StatusMessageLifetime = time.Second * 2
				m.userGlobal.renderEmoji = false
				cmds = append(cmds, m.list.NewStatusMessage("Emoji rendering disabled."))
			} else {
				m.list.StatusMessageLifetime = time.Second * 2
				m.userGlobal.renderEmoji = true
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

func (lm lobbyModel) View() string {
	if lm.showFH {
		return lm.fullHelp.View()
	}
	return docStyle.Render(lm.list.View())
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
			newItems := lm.userGlobal.rh.lobbyRequest()
			return itemsMsg{
				items: newItems,
			}
		})
	}

	return lm, nil
}

func (m *lobbyModel) joinGame(title string) tea.Cmd {
	return func() tea.Msg {
		gameId := gameId{GameId: title}
		if m.userGlobal.rh.joinGameRequest(gameId) {
			return joinGameMsg{
				gameId: gameId,
			}
		} else {
			return nil
		}
	}
}

type v1LobbyModel struct {
	model lobbyModel
}

func (m v1LobbyModel) Init() v1tea.Cmd {
	return m.Init()
}

func (m v1LobbyModel) Update(msg v1tea.Msg) (v1tea.Model, v1tea.Cmd) {
	return m.Update(msg)
}

func (m v1LobbyModel) View() string {
	return m.View()
}
