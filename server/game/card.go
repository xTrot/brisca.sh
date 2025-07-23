package game

import (
	"fmt"
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

type Card struct {
	Num   int
	Val   int
	Score int

	EmojiSuit string
	CharSuit  string

	// The original string from the server
	SuitString string
}

func NewCard(cardString string) Card {
	var card Card

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
		card.EmojiSuit = "🪙"
	case "COPA":
		card.EmojiSuit = "🏆"
	case "BASTO":
		card.EmojiSuit = "🪵"
	case "ESPADA":
		card.EmojiSuit = "⚔️"
	}
	switch suitString {
	case "ORO":
		card.CharSuit = "Or"
	case "COPA":
		card.CharSuit = "Co"
	case "BASTO":
		card.CharSuit = "Ba"
	case "ESPADA":
		card.CharSuit = "Es"
	}

	card.Num = CARDS_WITHOUT_SKIP[index][CARD_NUMBER_INDEX]
	card.Val = CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX]
	card.Score = CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX]

	card.SuitString = suitString

	return card
}

func (m *Card) RenderCard(renderEmoji bool) string {
	if renderEmoji {
		return fmt.Sprintf("[%s:%2d]", m.EmojiSuit, m.Num)
	} else {
		return fmt.Sprintf("[%s:%2d]", m.CharSuit, m.Num)
	}
}

func NewBottomCard(c Card) Card {
	// only a 2 of the same suit could do this
	swapNum := 2
	index := swapNum - 1
	return Card{
		EmojiSuit: c.EmojiSuit,
		CharSuit:  c.CharSuit,
		Num:       swapNum,
		Val:       CARDS_WITHOUT_SKIP[index][CARD_VALUE_INDEX],
		Score:     CARDS_WITHOUT_SKIP[index][CARD_SCORE_INDEX],
	}
}
