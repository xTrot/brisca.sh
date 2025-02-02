package main

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	wrUpdateInterval = time.Duration(time.Millisecond * 200)
)

var (
	wrDocStyle = lipgloss.NewStyle().Margin(1, 2)
)

type wrKeyMap struct {
	ready      key.Binding
	start      key.Binding
	changeTeam key.Binding
	leave      key.Binding
	spectate   key.Binding
}

func newWrKeyMap() *wrKeyMap {
	return &wrKeyMap{
		ready: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "ready"),
		),
		start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		changeTeam: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "change team"),
		),
		leave: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "leave"),
		),
		spectate: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "spectate"),
		),
	}
}

type waitingRoomModel struct {
	wr             waitingRoom
	list           list.Model
	keys           *wrKeyMap
	descDelegate   list.DefaultDelegate
	noDescDelegate list.DefaultDelegate
}

func newWaitingRoom() waitingRoomModel {
	var (
		listKeys = newWrKeyMap()
	)
	noDescDelegate := list.NewDefaultDelegate()
	noDescDelegate.ShowDescription = false
	descDelegate := list.NewDefaultDelegate()
	descDelegate.ShowDescription = true
	wrm := waitingRoomModel{
		wr:             waitingRoom{},
		list:           list.New([]list.Item{}, noDescDelegate, 0, 0),
		keys:           listKeys,
		descDelegate:   descDelegate,
		noDescDelegate: noDescDelegate,
	}

	wrm.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ready,
			listKeys.start,
			listKeys.changeTeam,
			listKeys.leave,
			listKeys.spectate,
		}
	}

	wrm.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ready,
			listKeys.start,
			listKeys.changeTeam,
			listKeys.leave,
		}
	}
	wrm.list.Title = "User " + username
	wrm.list.DisableQuitKeybindings()
	wrm.list.SetFilteringEnabled(false)
	wrm.list.SetShowStatusBar(false)
	wrm.list.SetShowPagination(false)
	return wrm
}

func (m waitingRoomModel) Init() tea.Cmd {
	return tea.Batch(every(wrUpdateInterval), tea.WindowSize())
}

func (m waitingRoomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case updateWRMsg:
		m.wr = msg.wr
		cmd = m.list.SetItems(m.wr.items)
		if m.wr.teams {
			m.list.SetDelegate(m.descDelegate)
		}
		cmds = append(cmds, cmd)
		cmds = append(cmds, every(wrUpdateInterval))
		if msg.wr.Started {
			gs := newGSModel()
			return gs, gs.Init()
		}

	case startGameMsg:
		gs := newGSModel()
		return gs, gs.Init()

	case leaveGameMsg:
		lobby := newLobby()
		lobby.list.Title = "User: " + username
		cmds = append(cmds, lobby.Init())
		lm, cmd := lobby.Update(msg)
		cmds = append(cmds, cmd)
		return lm, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.keys.ready):
			cmd = m.readyToggle()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.start):
			cmd = m.startGame()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.spectate):
			cmd = changeTeam(true)
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.changeTeam):
			cmd = changeTeam(false)
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.leave):
			cmd = m.leaveGame()
			cmds = append(cmds, cmd)

		}

	}

	list, cmd := m.list.Update(msg)
	m.list = list
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m waitingRoomModel) View() string {
	return docStyle.Render(m.list.View())
}

func every(interval time.Duration) tea.Cmd {
	return tea.Every(interval, updateWaitingRoom)
}

type updateWRMsg struct {
	wr waitingRoom
}

type readyToggleMsg struct{}

func (m waitingRoomModel) readyToggle() tea.Cmd {
	return func() tea.Msg {
		if readyRequest() {
			return readyToggleMsg{}
		} else {
			return nil
		}
	}
}

type startGameMsg struct{}

func (m waitingRoomModel) startGame() tea.Cmd {
	return func() tea.Msg {
		if startGameRequest() {
			return startGameMsg{}
		} else {
			return nil
		}
	}
}

type leaveGameMsg struct{}

func (m waitingRoomModel) leaveGame() tea.Cmd {
	return func() tea.Msg {
		if leaveGameRequest() {
			return leaveGameMsg{}
		} else {
			return nil
		}
	}
}

func updateWaitingRoom(t time.Time) tea.Msg {
	newWR := waitingRoomRequest()

	return updateWRMsg{
		wr: newWR,
	}
}

type changedTeamMsg bool

func changeTeam(spectator bool) tea.Cmd {
	return func() tea.Msg {
		success := changedTeamMsg(changeTeamRequest(spectator))
		return success
	}
}
