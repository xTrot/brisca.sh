package main

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func newItemDelegate(keys *delegateKeyMap, lm *lobbyModel) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		if i, ok := m.SelectedItem().(game); ok {
			title = i.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				var cmds []tea.Cmd
				cmd := m.NewStatusMessage(statusMessageStyle("You chose " + title))
				cmds = append(cmds, cmd)
				cmd = lm.joinGame(title)
				cmds = append(cmds, cmd)
				return tea.Sequence(cmds...)
			}
		}

		return nil
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}

type joinGameMsg struct {
	gameId gameId
}
