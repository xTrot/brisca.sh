package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	suitCardSwappedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("226"))
)

type tableModel struct {
	deckSize    int
	suitCard    card
	cardsInPlay []card

	suitCardStyle lipgloss.Style
}

func newTableModel() tableModel {
	return tableModel{
		deckSize: 40,
		suitCard: newCard("BASTO:3"),
		cardsInPlay: []card{
			newCard("ESPADA:1"),
			newCard("ORO:5"),
			newCard("COPA:10"),
		},
		suitCardStyle: lipgloss.NewStyle(),
	}
}

func (tm tableModel) Init() tea.Cmd {
	return nil
}

func (tm tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg.(type) {
	case swapBottomCardPayload:
		tm.suitCardStyle = suitCardSwappedStyle
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
	return fmt.Sprintf("Table:\n  Deck: %d\n  Suit Card: %s\n\n  In Play:\n  %s",
		tm.deckSize,
		tm.suitCardStyle.Render(tm.suitCard.renderCard()),
		tm.renderCardsInPlay(),
	)
}
