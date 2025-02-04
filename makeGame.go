package main

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

var (
	confirm bool = false
)

type makeGameModel struct {
	form       *huh.Form // huh.Form is just a tea.Model
	nextView   tea.Model
	userGlobal *userGlobal
}

func newMakeGame(nv tea.Model, userGlobal *userGlobal) makeGameModel {
	return makeGameModel{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("gameType").
					Options(huh.NewOptions("public", "private", "solo")...).
					Title("Choose a game type:"),

				huh.NewSelect[int]().
					Key("maxPlayers").
					Options(huh.NewOptions(2, 3, 4)...).
					Title("Max Players:"),

				huh.NewSelect[bool]().
					Key("swapBottomCard").
					Options(huh.NewOptions(true, false)...).
					Title("Is \"swapBottomCard\" allowed:"),

				huh.NewConfirm().
					Title("Are you sure?").
					Affirmative("Yes!").
					Negative("No.").
					Value(&confirm),
			),
		),
		nextView:   nv,
		userGlobal: userGlobal,
	}
}

func (m makeGameModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m makeGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ...

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		gc := gameConfig{
			GameType:       m.form.GetString("gameType"),
			MaxPlayers:     m.form.GetInt("maxPlayers"),
			SwapBottomCard: m.form.GetBool("swapBottomCard"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()

		game := m.userGlobal.rh.makeGameRequest(gc)

		err := spinner.New().
			Type(spinner.Line).
			Title("Making your Game...").
			Context(ctx).
			Run()

		if err != nil {
			log.Fatal(err)
		}

		wrm := newWaitingRoom(m.userGlobal)
		wrm.list.Title = "GameID: " + game.GameId
		cmd = wrm.Init()
		return wrm, cmd
	}

	return m, cmd
}

func (m makeGameModel) View() string {
	return m.form.View()
}
