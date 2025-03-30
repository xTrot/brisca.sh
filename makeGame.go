package main

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v2tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

type makeGameModel struct {
	form       *huh.Form // huh.Form is just a tea.Model
	nextView   tea.Model
	userGlobal userGlobal
	confirm    *bool

	helpMd MarkdownModel
	showMd bool
}

func newMakeGame(nv tea.Model, userGlobal userGlobal) makeGameModel {
	var confirm bool
	return makeGameModel{
		confirm: &confirm,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("gameType").
					Options(huh.NewOptions(
						"public",
						"private",
						"solo",
					)...).
					Title("Choose a game type:"),

				huh.NewSelect[int]().
					Key("maxPlayers").
					Options(huh.NewOptions(2, 3, 4)...).
					Title("Max Players:"),

				huh.NewSelect[bool]().
					Key("swapBottomCard").
					Options(huh.NewOptions(true, false)...).
					Title("Enable Swap Life Card house rule:"),

				huh.NewConfirm().
					Title("Are you sure?").
					Affirmative("Yes!").
					Negative("No.").
					Value(&confirm),
			),
		),
		nextView:   nv,
		userGlobal: userGlobal,
		helpMd:     NewMarkdownModel(MakeGameHelp, true, ""),
	}
}

func newV2MakeGame(nv tea.Model, userGlobal userGlobal) v2MakeGameModel {
	m := newMakeGame(nv, userGlobal)
	return v2MakeGameModel{model: m}
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

		if !*m.confirm {
			return m.nextView, m.nextView.Init()
		}

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

		wrm := newV1WaitingRoom(m.userGlobal)
		wrm.model.list.Title = "GameID: " + game.GameId
		cmd = wrm.Init()
		return wrm, cmd
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c":
			return m, tea.Quit
		case "H":
			m.showMd = !m.showMd
		}
	}

	return m, cmd
}

func (m makeGameModel) View() string {
	if m.showMd {
		return m.helpMd.View()
	} else {
		return m.form.View()
	}
}

type v2MakeGameModel struct {
	model makeGameModel
}

func (m v2MakeGameModel) Init() v2tea.Cmd {
	return m.Init()
}

func (m v2MakeGameModel) Update(msg v2tea.Msg) (v2tea.Model, v2tea.Cmd) {
	return m.Update(msg)
}

func (m v2MakeGameModel) View() string {
	return m.View()
}
