package main

import (
	"encoding/json"

	"github.com/charmbracelet/log"
)

// enum ActionType {
// 	GAME_CONFIG,
// 	GAME_STARTED,
// 	FIRST_PLAYER_SELECTED,
// 	BOTTOM_CARD_SELECTED,
// 	GRACE_PERIOD_ENDED,
// 	SWAP_BOTTOM_CARD,
// 	CARD_DRAWN,
// 	CARD_PLAYED,
// 	TURN_WON,
// 	GAME_WON
// }

type gameConfigPayload gameConfig

type gameStartedPayload struct {
	Seats []seat `json:"seats"`
}

type seat struct {
	Seat     int    `json:"seat"`
	Username string `json:"username"`
}

type firstPlayerSelectedPayload struct {
	Seat int `json:"seat"`
}

type bottomCardSelectedPayload struct {
	BottomCard string `json:"bottomCard"`
}

type cardDrawnPayload struct {
	Seat int `json:"seat"`
}

type cardPlayedPayload struct {
	Seat  int    `json:"seat"`
	Index int    `json:"index"`
	Card  string `json:"card"`
}

type turnWonPayload struct {
	Seat int `json:"seat"`
}

type gameWonPayload struct {
	Seat int    `json:"seat"`
	Team string `json:"team"`
}

type action struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload interface{}

type innerAction action

func (a *action) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*innerAction)(a))
	if err != nil {
		return err
	}

	switch a.Type {
	case "GAME_CONFIG":
		gameConfig := &gameConfigPayload{}
		err := json.Unmarshal(b, gameConfig)
		if err != nil {
			return err
		}
		a.Payload = gameConfig
	case "GAME_STARTED":
		gameStarted := &gameStartedPayload{}
		err := json.Unmarshal(b, gameStarted)
		if err != nil {
			return err
		}
		a.Payload = gameStarted
	case "FIRST_PLAYER_SELECTED":
		playerSelected := &firstPlayerSelectedPayload{}
		err := json.Unmarshal(b, playerSelected)
		if err != nil {
			return err
		}
		a.Payload = playerSelected
	case "BOTTOM_CARD_SELECTED":
		bottomCard := &bottomCardSelectedPayload{}
		err := json.Unmarshal(b, bottomCard)
		if err != nil {
			return err
		}
		a.Payload = bottomCard
	case "GRACE_PERIOD_ENDED":
		a.Payload = nil
	case "SWAP_BOTTOM_CARD":
		a.Payload = nil
	case "CARD_DRAWN":
		cardDrawn := &cardDrawnPayload{}
		err := json.Unmarshal(b, cardDrawn)
		if err != nil {
			return err
		}
		a.Payload = cardDrawn
	case "CARD_PLAYED":
		cardPlayed := &cardPlayedPayload{}
		err := json.Unmarshal(b, cardPlayed)
		if err != nil {
			return err
		}
		a.Payload = cardPlayed
	case "TURN_WON":
		turnWon := &turnWonPayload{}
		err := json.Unmarshal(b, turnWon)
		if err != nil {
			return err
		}
		a.Payload = turnWon
	case "GAME_WON":
		gameWon := &gameWonPayload{}
		err := json.Unmarshal(b, gameWon)
		if err != nil {
			return err
		}
		a.Payload = gameWon
	default:
		log.Errorf("action.UnmarshalJSON: unexpected type; type = %s)", a.Type)
	}

	return nil
}
