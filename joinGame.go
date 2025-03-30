package main

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v2tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/log"
)

type joinGameModel struct {
	form       *huh.Form // huh.Form is just a tea.Model
	nextView   tea.Model
	userGlobal userGlobal
	gameId     *string
	replay     bool
}

func newReplayGame(nv tea.Model, userGlobal userGlobal) joinGameModel {
	var gameId string
	return joinGameModel{
		gameId: &gameId,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What's the UUID of the Game you want to replay?").
					Value(&gameId),
			),
		),
		nextView:   nv,
		userGlobal: userGlobal,
		replay:     true,
	}
}

func newJoinGame(nv tea.Model, userGlobal userGlobal) joinGameModel {
	var gameId string
	return joinGameModel{
		gameId: &gameId,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What's the UUID of the Game you want to join?").
					Value(&gameId),
			),
		),
		nextView:   nv,
		userGlobal: userGlobal,
	}
}

func newV2ReplayGame(nv tea.Model, userGlobal userGlobal) v2JoinGameModel {
	m := newReplayGame(nv, userGlobal)
	return v2JoinGameModel{m}
}

func newV2JoinGame(nv tea.Model, userGlobal userGlobal) v2JoinGameModel {
	m := newJoinGame(nv, userGlobal)
	return v2JoinGameModel{m}
}

func (m joinGameModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m joinGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ...

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		var gameId gameId
		gameId.GameId = *m.gameId

		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()

		err := spinner.New().
			Type(spinner.Line).
			Title("Finding your Game...").
			Context(ctx).
			Run()

		if err != nil {
			log.Fatal(err)
		}

		if m.replay {
			replay := m.userGlobal.rh.replayRequest(gameId)

			if replay != nil {
				rgs := newV1ReplayGSModel(m.userGlobal, replay)
				return rgs, rgs.Init()
			} else {
				return m.nextView, m.nextView.Init()
			}
		}

		joined := m.userGlobal.rh.joinGameRequest(gameId)

		if joined {
			wrm := newV1WaitingRoom(m.userGlobal)
			wrm.model.list.Title = "GameID: " + gameId.GameId
			cmd = wrm.Init()
			return wrm, cmd
		} else {
			return m.nextView, m.nextView.Init()
		}

	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c":
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m joinGameModel) View() string {
	return m.form.View()
}

type v2JoinGameModel struct {
	model joinGameModel
}

func (m v2JoinGameModel) Init() v2tea.Cmd {
	return m.Init()
}

func (m v2JoinGameModel) Update(msg v2tea.Msg) (v2tea.Model, v2tea.Cmd) {
	return m.Update(msg)
}

func (m v2JoinGameModel) View() string {
	return m.View()
}
