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
	for i := range pileSize {
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
		name:      "Username",
		score:     0,
		scorePile: []card{},
		boxX:      2, // Coords for mySeat
		boxY:      1, // Coords for mySeat
	}
}

func (pm playerModel) Init() tea.Cmd {
	return nil
}

func (pm playerModel) Update(msg tea.Msg) (playerModel, tea.Cmd) {
	return pm, nil
}

func (pm playerModel) View(x, y int) string {
	const twoNewLines int = 2
	remainingY := y - twoNewLines
	return fmt.Sprintf("%s Score: %d\n  Score Pile:\n  %s",
		pm.name, pm.score, pm.renderScorePile(x, remainingY))
}

func (pm playerModel) renderScorePile(x, y int) string {
	// I only have space for 5x4 cards
	cardLength := 6
	const paddingBothSides int = 4
	maxRows := max(y, 0)
	maxCols := max((x-paddingBothSides)/cardLength-1, 0)
	maxSP := maxRows * maxCols
	spString := ""
	spSize := len(pm.scorePile)
	reverseStart := spSize - 1
rowLoop:
	for i := range maxRows {
		indexBase := maxCols * i
		for j := range maxCols {
			index := indexBase + j
			if !(index < spSize) {
				break rowLoop // Breaks out of both loops
			}
			spString += renderCard(pm.scorePile[reverseStart-index])
		}
		spString += "\n  " // paddingBothSides
	}
	if maxSP != 0 && spSize == maxSP {
		spString = spString[:len(spString)-3] // remove last spacing
	} else if spSize > maxSP {
		spString = spString[:len(spString)-3] // remove last spacing
		spString += "..."
	}
	return spString
}
