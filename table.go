package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type tableModel struct {
	deckSize    int
	suitCard    card
	cardsInPlay []card
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
	}
}

func (tm tableModel) Init() tea.Cmd {
	return nil
}

func (tm tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		tm.deckSize, tm.suitCard.renderCard(), tm.renderCardsInPlay())
}
