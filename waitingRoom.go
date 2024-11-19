package main

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	wrUpdateInterval = time.Duration(time.Millisecond * 200)
)

type wrKeyMap struct {
	ready key.Binding
	start key.Binding
	leave key.Binding
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
		leave: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "leave"),
		),
	}
}

type waitingRoomModel struct {
	wr   waitingRoom
	list list.Model
	keys *wrKeyMap
}

func newWaitingRoom() waitingRoomModel {
	var (
		listKeys = newWrKeyMap()
	)
	wrm := waitingRoomModel{
		wr:   waitingRoom{},
		list: list.New([]list.Item{}, list.DefaultDelegate{}, 0, 0),
		keys: listKeys,
	}
	wrm.list.Title = "User " + username
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

	case startGameMsg:
		panic("startGame not implemented yet.")

	case leaveGameMsg:
		lobby := newLobby()
		lobby.list.Title = "User: " + username
		cmds = append(cmds, lobby.Init())
		lm, cmd := lobby.Update(msg)
		cmds = append(cmds, cmd)
		return lm, tea.Batch(cmds...)

	case updateWRMsg:
		m.wr = msg.wr
		cmd = m.list.SetItems(m.wr.items)
		cmds = append(cmds, cmd)
		cmds = append(cmds, every(wrUpdateInterval))

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
