package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type CheatSheetModel struct {
	Text string
}

func NewCheatSheetModel() CheatSheetModel {
	return CheatSheetModel{
		Text: CheatSheet,
	}
}

func (m CheatSheetModel) Init() tea.Cmd {
	return nil
}

func (m CheatSheetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m CheatSheetModel) View() string {
	out, err := glamour.Render(m.Text, "dracula")
	if err != nil {
		log.Fatal(err.Error())
		panic("Markdown render panic.")
	}
	return out
}
