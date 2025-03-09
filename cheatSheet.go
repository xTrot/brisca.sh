package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
)

type CheatSheetModel struct {
	Text  string
	Style lipgloss.Style
}

func NewCheatSheetModel() CheatSheetModel {
	return CheatSheetModel{
		Text:  CheatSheet,
		Style: style,
	}
}

func (m CheatSheetModel) Init() tea.Cmd {
	return nil
}

func (m CheatSheetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m CheatSheetModel) View() string {
	glamour.WithWordWrap(m.Style.GetWidth())
	out, err := glamour.Render(m.Text, "dracula")
	if err != nil {
		log.Fatal(err.Error())
		panic("Markdown render panic.")
	}
	return out
}
