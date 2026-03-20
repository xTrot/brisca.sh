package game

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

type GameConfigPayload struct {
	GameId         string `json:"gameId"`
	GameType       string `json:"gameType"`
	MaxPlayers     int    `json:"maxPlayers"`
	SwapBottomCard bool   `json:"swapBottomCard"`
}

type GameStartedPayload struct {
	Seats        []Seat `json:"seats"`
	StartingSeat int    `json:"startingSeat"`
}

type Seat struct {
	Seat     int    `json:"seat"`
	Username string `json:"username"`
}

type GracePeriodEndedPayload struct{}

type SwapBottomCardPayload struct{}

type BottomCardSelectedPayload struct {
	CardString string `json:"bottomCard"`
	Card       Card
}

type CardDrawnPayload struct {
	Seat int `json:"seat"`
}

type CardPlayedPayload struct {
	Seat       int    `json:"seat"`
	Index      int    `json:"index"`
	CardString string `json:"card"`
	Card       Card
}

type TurnWonPayload struct {
	Seat int `json:"seat"`
}

type GameWonPayload struct {
	Seat int    `json:"seat"`
	Team string `json:"team"`
}

type SeatAfkPayload struct {
	Seat int `json:"seat"`
}

type SeatNotAfkPayload struct {
	Seat int `json:"seat"`
}

// Client side action payloads
// ============================================================================
type TurnSwitchPayload struct{}

type UndefinedActionPayload struct{}

// ============================================================================

type Action struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload any

type innerAction struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

func (a Action) ProcessAction(myTurn, gameOver bool, mySeat int) tea.Cmd {
	return func() tea.Msg {
		slow := COMMON_WAIT
		switch payload := a.Payload.(type) {
		case GameConfigPayload:
			slow = 0
		case GameStartedPayload:
			slow = 0
		case BottomCardSelectedPayload:
			slow = 0
		case GracePeriodEndedPayload:
		case SwapBottomCardPayload:
			if myTurn && !gameOver {
				slow = 0
			}
		case CardDrawnPayload:
			slow = time.Millisecond * 200
		case CardPlayedPayload:
			if payload.Seat == mySeat && !gameOver {
				slow = 0
			}
		case TurnWonPayload:
		case GameWonPayload:

		// Client side actions
		case TurnSwitchPayload:
			slow = time.Millisecond * 200

		}
		time.Sleep(slow)
		return a.Payload
	}
}

func (a Action) String() string {
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

func (a *Action) UnmarshalJSON(b []byte) error {
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
		gameConfig := GameConfigPayload{}
		err := json.Unmarshal(payloadBytes, &gameConfig)
		if err != nil {
			return err
		}
		a.Payload = gameConfig
	case "GAME_STARTED":
		gameStarted := GameStartedPayload{}
		err := json.Unmarshal(payloadBytes, &gameStarted)
		if err != nil {
			return err
		}
		a.Payload = gameStarted
	case "BOTTOM_CARD_SELECTED":
		bottomCard := BottomCardSelectedPayload{}
		err := json.Unmarshal(payloadBytes, &bottomCard)
		if err != nil {
			return err
		}
		bottomCard.Card = NewCard(bottomCard.CardString)
		a.Payload = bottomCard
	case "GRACE_PERIOD_ENDED":
		a.Payload = GracePeriodEndedPayload{}
	case "SWAP_BOTTOM_CARD":
		a.Payload = SwapBottomCardPayload{}
	case "CARD_DRAWN":
		cardDrawn := CardDrawnPayload{}
		err := json.Unmarshal(payloadBytes, &cardDrawn)
		if err != nil {
			return err
		}
		a.Payload = cardDrawn
	case "CARD_PLAYED":
		cardPlayed := CardPlayedPayload{}
		err := json.Unmarshal(payloadBytes, &cardPlayed)
		if err != nil {
			return err
		}
		cardPlayed.Card = NewCard(cardPlayed.CardString)
		a.Payload = cardPlayed
	case "TURN_WON":
		turnWon := TurnWonPayload{}
		err := json.Unmarshal(payloadBytes, &turnWon)
		if err != nil {
			return err
		}
		a.Payload = turnWon
	case "GAME_WON":
		gameWon := GameWonPayload{}
		err := json.Unmarshal(payloadBytes, &gameWon)
		if err != nil {
			return err
		}
		a.Payload = gameWon
	case "SEAT_AFK":
		seat := SeatAfkPayload{}
		err := json.Unmarshal(payloadBytes, &seat)
		if err != nil {
			return err
		}
		a.Payload = seat
	case "SEAT_NOT_AFK":
		seat := SeatNotAfkPayload{}
		err := json.Unmarshal(payloadBytes, &seat)
		if err != nil {
			return err
		}
		a.Payload = seat
	default:
		logger.Error("action.UnmarshalJSON: unexpected", "type", a.Type)
		a.Payload = UndefinedActionPayload{}
	}

	return nil
}
