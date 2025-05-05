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
	renderEmoji     bool
}

func newTableModel(renderEmoji bool) tableModel {
	return tableModel{
		deckSize:   40,
		bottomCard: newCard("BASTO:3"),
		cardsInPlay: []card{
			newCard("ESPADA:1"),
			newCard("ORO:5"),
			newCard("COPA:10"),
		},
		bottomCardStyle: lipgloss.NewStyle(),
		renderEmoji:     renderEmoji,
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

func (tm tableModel) renderCardsInPlay(width int, height int) string {
	if height <= 2 {
		cip := ""
		switch {
		case len(tm.cardsInPlay) > 0:
			cip += tm.cardsInPlay[0].renderCard(tm.renderEmoji)
			fallthrough
		case len(tm.cardsInPlay) > 1:
			cip += "\n  "
			for i := 1; i < len(tm.cardsInPlay); i++ {
				cip += tm.cardsInPlay[i].renderCard(tm.renderEmoji)
			}
		}
		return cip
	}
	cardLength := 7
	padding := "\n  " // padding 2 since new line doesn't count
	const paddingBothSides int = 4
	maxRows := 2
	maxCols := max((width-paddingBothSides)/cardLength, 0)
	maxCIP := maxRows * maxCols
	cip := padding
	cipSize := len(tm.cardsInPlay)
rowLoop:
	for i := range maxRows {
		indexBase := maxCols * i
		for j := range maxCols {
			index := indexBase + j
			if !(index < cipSize) {
				break rowLoop
			}
			cip += tm.cardsInPlay[index].renderCard(tm.renderEmoji)
		}
		cip += padding
	}
	if cipSize >= maxCIP {
		cip = cip[:len(cip)-3]
	}
	return cip
}

func (tm tableModel) View(width int, height int) string {
	return fmt.Sprintf("Table:\n  Deck: %d\n  Life Card: %s\n  In Play: %s",
		tm.deckSize,
		tm.bottomCardStyle.Render(tm.bottomCard.renderCard(tm.renderEmoji)),
		tm.renderCardsInPlay(width, height-3), // 3 new lines above.
	)
}
