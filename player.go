package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	SEAT_BASED_BOXES_2P = [][]int{
		{2, 1},
		{0, 1},
	}
	SEAT_BASED_BOXES_3P = [][]int{
		{2, 1},
		{0, 1},
		{1, 2},
	}
	SEAT_BASED_BOXES_4P = [][]int{
		{2, 1},
		{1, 0},
		{0, 1},
		{1, 2},
	}
)

type playerModel struct {
	name      string
	score     int
	scorePile []card
	handSize  int
	boxX      int
	boxY      int
}

func (pm playerModel) UpdateScore() int {
	pileSize := len(pm.scorePile)
	score := 0
	for i := 0; i < pileSize; i++ {
		score += pm.scorePile[i].score
	}
	return score
}

func newPlayerModelFromSeat(s seat) playerModel {
	return playerModel{
		name:      s.Username,
		score:     0,
		scorePile: []card{},
		handSize:  3,
	}
}

func newPlayerModel() playerModel {
	return playerModel{
		name:  "Username",
		score: 0,
		scorePile: []card{
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
			newCard("ESPADA:3"),
			newCard("BASTO:2"),
			newCard("COPA:6"),
			newCard("COPA:7"),
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
	reverseStart := spSize - 1
rowLoop:
	for i := 0; i < maxRows; i++ {
		indexBase := maxCols * i
		for j := 0; j < maxCols; j++ {
			index := indexBase + j
			if !(index < spSize) {
				break rowLoop // Breaks out of both loops
			}
			spString += renderCard(pm.scorePile[reverseStart-index])
		}
		spString += "\n  " // spacing
	}
	if spSize == maxSP {
		spString = spString[:len(spString)-3] // remove last spacing
	} else if spSize > maxSP {
		spString = spString[:len(spString)-3] // remove last spacing
		spString += " ..."
	}
	return spString
}
