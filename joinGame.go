package main

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
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

func (m joinGameModel) Init() tea.Cmd {
	return func() tea.Msg {
		function := m.form.Init()
		if function != nil {
			return function()
		}
		return nil
	}
}

func (m joinGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ...

	var cmd tea.Cmd
	form, v1cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if v1cmd != nil {
		cmd = func() tea.Msg {
			return v1cmd()
		}
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
				rgs := newReplayGSModel(m.userGlobal, replay)
				return rgs, rgs.Init()
			} else {
				return m.nextView, m.nextView.Init()
			}
		}

		joined := m.userGlobal.rh.joinGameRequest(gameId)

		if joined {
			wrm := newWaitingRoom(m.userGlobal)
			wrm.list.Title = "GameID: " + gameId.GameId
			cmd = func() tea.Msg {
				funtion := wrm.Init()
				return funtion()
			}
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
