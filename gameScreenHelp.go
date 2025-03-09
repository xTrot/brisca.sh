package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type gameScreenKeyMap struct {
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	One      key.Binding
	Two      key.Binding
	Three    key.Binding
	Swap     key.Binding
	Help     key.Binding
	Quit     key.Binding
	showSwap bool
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k gameScreenKeyMap) ShortHelp() []key.Binding {
	if k.showSwap {
		return []key.Binding{k.Left, k.Right, k.Enter, k.Swap}
	} else {
		return []key.Binding{k.Left, k.Right, k.Enter}
	}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k gameScreenKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Enter, k.One, k.Two, k.Three, k.Swap, k.Help, k.Quit}, // second column
	}
}

var gameScreenKeys = gameScreenKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "play card"),
	),
	One: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "play card 1"),
	),
	Two: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "play card 2"),
	),
	Three: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "play card 3"),
	),
	Swap: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "swap suit card"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

type gameScreenHelpModel struct {
	keys gameScreenKeyMap
	help help.Model
}

func newGSHelp() gameScreenHelpModel {
	return gameScreenHelpModel{
		keys: gameScreenKeys,
		help: help.New(),
	}
}

func (hm gameScreenHelpModel) Init() tea.Cmd {
	return nil
}

func (hm gameScreenHelpModel) Update(msg tea.Msg) (gameScreenHelpModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, hm.keys.Help):
			hm.help.ShowAll = !hm.help.ShowAll
		}
	}
	return hm, nil
}

func (hm gameScreenHelpModel) View() string {

	return hm.help.View(hm.keys)
}
