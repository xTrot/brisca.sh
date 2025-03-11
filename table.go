package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bottomCardSwappedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("226"))
)

type tableModel struct {
	deckSize    int
	bottomCard  card
	cardsInPlay []card

	bottomCardStyle lipgloss.Style
}

func newTableModel() tableModel {
	return tableModel{
		deckSize:   40,
		bottomCard: newCard("BASTO:3"),
		cardsInPlay: []card{
			newCard("ESPADA:1"),
			newCard("ORO:5"),
			newCard("COPA:10"),
		},
		bottomCardStyle: lipgloss.NewStyle(),
	}
}

func (tm tableModel) Init() tea.Cmd {
	return nil
}

func (tm tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg.(type) {
	case swapBottomCardPayload:
		tm.bottomCardStyle = bottomCardSwappedStyle
	}

	return tm, nil
}

func (tm tableModel) renderCardsInPlay() string {
	cip := ""
	for i := range len(tm.cardsInPlay) {
		cip += fmt.Sprintf("%s ", tm.cardsInPlay[i].renderCard())
	}
	return cip
}

func (tm tableModel) View() string {
	return fmt.Sprintf("Table:\n  Deck: %d\n  Life Card: %s\n\n  In Play:\n  %s",
		tm.deckSize,
		tm.bottomCardStyle.Render(tm.bottomCard.renderCard()),
		tm.renderCardsInPlay(),
	)
}
