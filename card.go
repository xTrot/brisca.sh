package main

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
)

type card struct {
	suit  string
	num   int
	val   int
	score int
}

func newCard(suitString string, cardNum int) card {
	var card card

	index := cardNum - 1

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
