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

func (tm tableModel) Update(msg tea.Msg) (tableModel, tea.Cmd) {

	switch msg.(type) {
	case swapBottomCardPayload:
		tm.bottomCardStyle = bottomCardSwappedStyle
	}

	return tm, nil
}

func (tm tableModel) renderCardsInPlay(width int) string {
	cardLength := 7
	const paddingBothSides int = 4
	maxRows := 2
	maxCols := max((width-paddingBothSides)/cardLength-1, 0)
	maxCIP := maxRows * maxCols
	cip := ""
	cipSize := len(tm.cardsInPlay)
rowLoop:
	for i := range maxRows {
		indexBase := maxCols * i
		for j := range maxCols {
			index := indexBase + j
			if !(index < cipSize) {
				break rowLoop
			}
			cip += fmt.Sprintf("%s ", tm.cardsInPlay[index].renderCard())
		}
		cip += "\n  "
	}
	if cipSize >= maxCIP {
		cip = cip[:len(cip)-3]
	}
	return cip
}

func (tm tableModel) View(width int) string {
	return fmt.Sprintf("Table:\n  Deck: %d\n  Life Card: %s\n\n  In Play:\n  %s",
		tm.deckSize,
		tm.bottomCardStyle.Render(tm.bottomCard.renderCard()),
		tm.renderCardsInPlay(width),
	)
}
