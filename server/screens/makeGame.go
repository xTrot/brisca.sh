package screens

import (
	"time"

	"brisca.sh/server/embedded"
	"brisca.sh/server/requests"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type makeGameModel struct {
	form      *huh.Form // huh.Form is just a tea.Model
	nextView  tea.Model
	usc       UserScreenContext
	confirm   *bool
	waiting   bool
	waitStyle lipgloss.Style
	spinner   spinner.Model

	helpMd MarkdownModel
	showMd bool
}

func newMakeGame(nv tea.Model, usc UserScreenContext) makeGameModel {
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
		nextView: nv,
		usc:      usc,
		helpMd:   NewMarkdownModel(embedded.MakeGameHelp, true, ""),
		waiting:  false,
		waitStyle: usc.Renderer().NewStyle().
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(
				usc.Renderer().NewStyle().
					AlignHorizontal(lipgloss.Center).
					AlignVertical(lipgloss.Center),
			),
		),
	}
}

func (m makeGameModel) Init() tea.Cmd {
	return tea.Batch(
		m.form.Init(),
		m.spinner.Tick,
		m.usc.LastWindowSizeReplay(),
	)
}

func (m makeGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ...

	var cmds []tea.Cmd
	quit, cmd := HasRetiredUser(&m.usc)
	if cmd != nil {
		return quit, cmd
	}

	if time.Now().After(m.usc.ReqHandler().RefreshBy) {
		logger.Debug("Idle Disconnect")
		return NewModel("Idle Disconnect")
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	cmds = append(cmds, cmd)

	if m.form.State == huh.StateCompleted && !m.waiting {

		if !*m.confirm {
			return m.nextView, m.nextView.Init()
		}

		gc := requests.GameConfig{
			GameType:       m.form.GetString("gameType"),
			MaxPlayers:     m.form.GetInt("maxPlayers"),
			SwapBottomCard: m.form.GetBool("swapBottomCard"),
		}

		cmds = append(cmds, m.makeGame(gc))
		m.waiting = true

	}

	switch msg := msg.(type) {

	case requests.NewGame:

		// Null Check
		if msg.GameId == (requests.NewGame{}).GameId {
			return m.nextView, m.nextView.Init()
		}

		logger.Debug("Making game successful", "game", msg)
		m.usc.ReqHandler().SetGameServer(msg.GameServer)
		wrm := newWaitingRoom(m.usc)
		wrm.list.Title = "GameID: " + msg.GameId
		return wrm, wrm.Init()

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl-c":
			return m, tea.Quit
		case "H":
			m.showMd = !m.showMd

		}

	case tea.WindowSizeMsg:
		m.usc.SetWindowsSize(msg)
		m.waitStyle = m.waitStyle.
			Height(msg.Height).
			Width(msg.Width)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)

}

func (m makeGameModel) makeGame(gc requests.GameConfig) tea.Cmd {
	return func() tea.Msg {
		logger.Debug("makeGame request", "gc", gc)
		time.Sleep(time.Second)
		return m.usc.ReqHandler().MakeGameRequest(gc)
	}
}

func (m makeGameModel) View() string {
	if m.showMd {
		return m.helpMd.View()
	} else if m.waiting {
		return m.waitStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.spinner.View(),
				" Making Game",
			),
		)
	} else {
		return m.form.View()
	}
}
