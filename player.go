package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type playerModel struct {
	name      string
	score     int
	scorePile []card
}

func newplayerModel() playerModel {
	return playerModel{
		name:  "Username",
		score: 0,
		scorePile: []card{
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
			newCard("ESPADA", 3),
			newCard("BASTO", 2),
			newCard("COPA", 6),
			newCard("COPA", 7),
		},
	}
}

func (pm playerModel) Init() tea.Cmd {
	return nil
}

func (pm playerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return pm, nil
}

func (pm playerModel) View() string {
	return fmt.Sprintf("%s Score: %d\n  Score Pile:\n  %s", pm.name, pm.score, pm.renderScorePile())
}

func (pm playerModel) renderScorePile() string {
	// I only have space for 5x4 cards
	maxRows := 5
	maxCols := 4
	maxSP := maxRows * maxCols
	spString := ""
	spSize := len(pm.scorePile)
rowLoop:
	for i := 0; i < maxRows; i++ {
		indexBase := maxCols * i
		for j := 0; j < maxCols; j++ {
			index := indexBase + j
			if !(index < spSize) {
				break rowLoop // Breaks out of both loops
			}
			spString += renderCard(pm.scorePile[index])
		}
		spString += "\n  " // spacing
	}
	if spSize > maxSP {
		spString = spString[:len(spString)-3] // remove last spacing
		spString += " ..."
	}
	return spString
}
