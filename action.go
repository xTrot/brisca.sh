package main

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

const (
	COMMON_WAIT time.Duration = time.Second
)

// enum ActionType {
// 	GAME_CONFIG,
// 	GAME_STARTED,
// 	BOTTOM_CARD_SELECTED,
// 	GRACE_PERIOD_ENDED,
// 	SWAP_BOTTOM_CARD,
// 	CARD_DRAWN,
// 	CARD_PLAYED,
// 	TURN_WON,
// 	GAME_WON,
// 	SEAT_AFK,
// 	SEAT_NOT_AFK,
// }

type gameConfigPayload struct {
	GameId         string `json:"gameId"`
	GameType       string `json:"gameType"`
	MaxPlayers     int    `json:"maxPlayers"`
	SwapBottomCard bool   `json:"swapBottomCard"`
}

type gameStartedPayload struct {
	Seats        []seat `json:"seats"`
	StartingSeat int    `json:"startingSeat"`
}

type seat struct {
	Seat     int    `json:"seat"`
	Username string `json:"username"`
}

type gracePeriodEndedPayload struct{}

type swapBottomCardPayload struct{}

type bottomCardSelectedPayload struct {
	BottomCard string `json:"bottomCard"`
	bottomCard card
}

type cardDrawnPayload struct {
	Seat int `json:"seat"`
}

type cardPlayedPayload struct {
	Seat  int    `json:"seat"`
	Index int    `json:"index"`
	Card  string `json:"card"`
	card  card
}

type turnWonPayload struct {
	Seat int `json:"seat"`
}

type gameWonPayload struct {
	Seat int    `json:"seat"`
	Team string `json:"team"`
}

type seatAfkPayload struct {
	Seat int `json:"seat"`
}

type seatNotAfkPayload struct {
	Seat int `json:"seat"`
}

// Client side action payloads
// ============================================================================
type turnSwitchPayload struct{}

type undefinedActionPayload struct{}

// ============================================================================

type action struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload any

type innerAction struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

func (a action) processAction(myTurn, gameOver bool, mySeat int) tea.Cmd {
	return func() tea.Msg {
		slow := COMMON_WAIT
		switch payload := a.Payload.(type) {
		case gameConfigPayload:
			slow = 0
		case gameStartedPayload:
			slow = 0
		case bottomCardSelectedPayload:
			slow = 0
		case gracePeriodEndedPayload:
		case swapBottomCardPayload:
			if myTurn && !gameOver {
				slow = 0
			}
		case cardDrawnPayload:
			slow = time.Millisecond * 200
		case cardPlayedPayload:
			if payload.Seat == mySeat && !gameOver {
				slow = 0
			}
		case turnWonPayload:
		case gameWonPayload:

		// Client side actions
		case turnSwitchPayload:
			slow = time.Millisecond * 200

		}
		time.Sleep(slow)
		return a.Payload
	}
}

func (a action) String() string {
	return fmt.Sprintf("{Type:%s Payload:%s}", a.Type, a.Payload)
}

func findOcurrence(bytes []byte, char byte, ocurrence int, dir int) int {
	found := -1

	startingFrom := -1
	if dir == 1 {
		startingFrom = 0
		max := len(bytes)
		for i := ocurrence; i > 0; i-- {
			for j := startingFrom; j < max; j++ {
				if char == bytes[j] {
					found = j
					startingFrom = j + 1
					break
				}
			}
		}
	} else if dir == -1 {
		startingFrom = len(bytes) - 1
		for i := ocurrence; i > 0; i-- {
			for j := startingFrom; j >= 0; j-- {
				if char == bytes[j] {
					found = j + 1
					startingFrom = j - 1
					break
				}
			}
		}
	} else {
		return -1
	}

	return found
}

func (a *action) UnmarshalJSON(b []byte) error {
	var ia innerAction
	err := json.Unmarshal(b, &ia)
	if err != nil {
		return err
	}

	a.Type = ia.Type
	a.Payload = ia.Payload

	to := findOcurrence(b, '{', 2, 1)
	from := findOcurrence(b, '}', 2, -1)
	if to == -1 || from == -1 {
		return nil
	}
	payloadBytes := b[to:from]

	switch a.Type {
	case "GAME_CONFIG":
		gameConfig := gameConfigPayload{}
		err := json.Unmarshal(payloadBytes, &gameConfig)
		if err != nil {
			return err
		}
		a.Payload = gameConfig
	case "GAME_STARTED":
		gameStarted := gameStartedPayload{}
		err := json.Unmarshal(payloadBytes, &gameStarted)
		if err != nil {
			return err
		}
		a.Payload = gameStarted
	case "BOTTOM_CARD_SELECTED":
		bottomCard := bottomCardSelectedPayload{}
		err := json.Unmarshal(payloadBytes, &bottomCard)
		if err != nil {
			return err
		}
		bottomCard.bottomCard = newCard(bottomCard.BottomCard)
		a.Payload = bottomCard
	case "GRACE_PERIOD_ENDED":
		a.Payload = gracePeriodEndedPayload{}
	case "SWAP_BOTTOM_CARD":
		a.Payload = swapBottomCardPayload{}
	case "CARD_DRAWN":
		cardDrawn := cardDrawnPayload{}
		err := json.Unmarshal(payloadBytes, &cardDrawn)
		if err != nil {
			return err
		}
		a.Payload = cardDrawn
	case "CARD_PLAYED":
		cardPlayed := cardPlayedPayload{}
		err := json.Unmarshal(payloadBytes, &cardPlayed)
		if err != nil {
			return err
		}
		cardPlayed.card = newCard(cardPlayed.Card)
		a.Payload = cardPlayed
	case "TURN_WON":
		turnWon := turnWonPayload{}
		err := json.Unmarshal(payloadBytes, &turnWon)
		if err != nil {
			return err
		}
		a.Payload = turnWon
	case "GAME_WON":
		gameWon := gameWonPayload{}
		err := json.Unmarshal(payloadBytes, &gameWon)
		if err != nil {
			return err
		}
		a.Payload = gameWon
	case "SEAT_AFK":
		a.Payload = seatAfkPayload{}
	case "SEAT_NOT_AFK":
		a.Payload = seatNotAfkPayload{}
	default:
		log.Errorf("action.UnmarshalJSON: unexpected type: type = %s", a.Type)
		a.Payload = undefinedActionPayload{}
	}

	return nil
}
