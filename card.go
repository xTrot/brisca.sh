package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

var (
	CARD_NUMBER_INDEX  = 0
	CARD_VALUE_INDEX   = 1
	CARD_SCORE_INDEX   = 2
	CARDS_WITHOUT_SKIP = [][]int{
		{1, 12, 11},
		{2, 1, 0},
		{3, 11, 10},
		{4, 2, 0},
		{5, 3, 0},
		{6, 4, 0},
		{7, 5, 0},
		{8, 6, 0},
		{9, 7, 0},
		{10, 8, 2},
		{11, 9, 3},
		{12, 10, 4},
	}
	SUITS = []string{
		"ORO",
		"COPA",
		"BASTO",
		"ESPADA",
	}
)

type card struct {
	suit  string
	num   int
	val   int
	score int

	// The original string from the server
	suitString string
}

func newCard(cardString string) card {
	var card card

	halves := strings.Split(cardString, ":")
	suitString := halves[0]
	num, err := strconv.Atoi(halves[1])
	if err != nil {
		log.Error("Error parsing str to int for hand request.")
		return card
	}

	index := num - 1

	switch suitString {
	case "ORO":
		card.suit = "🪙"
	case "COPA":
		card.suit = "🏆"
	case "BASTO":
		card.suit = "🪵"
	case "ESPADA":
		card.suit = "⚔️"
	}

	card.num = CARDS_WITHOUT_SKIP[index][CARD_NUMBER_INDEX]
	card.val = CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX]
	card.score = CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX]

	card.suitString = suitString

	return card
}

func (m *card) renderCard() string {
	return fmt.Sprintf("[%s:%2d]", m.suit, m.num)
}

func randCard() card {
	var card card

	index := rand.IntN(len(CARDS_WITHOUT_SKIP))

	suitString := SUITS[rand.IntN(len(SUITS))]

	switch suitString {
	case "ORO":
		card.suit = "🪙"
	case "COPA":
		card.suit = "🏆"
	case "BASTO":
		card.suit = "🪵"
	case "ESPADA":
		card.suit = "⚔️"
	}

	card.num = CARDS_WITHOUT_SKIP[index][CARD_NUMBER_INDEX]
	card.val = CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX]
	card.score = CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX]

	return card
}

func newBottomCard(c card) card {
	// only a 2 of the same suit could do this
	swapNum := 2
	index := swapNum - 1
	return card{
		suit:  c.suit,
		num:   swapNum,
		val:   CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX],
		score: CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX],
	}
}
