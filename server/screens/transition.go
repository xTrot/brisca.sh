package screens

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TransitionModel struct {
	usc UserScreenContext

	waitStyle  lipgloss.Style
	spinner    spinner.Model
	nextScreen tea.Model
	duration   time.Duration
	msg        string
}

func Transition(
	usc UserScreenContext,
	nextScreen tea.Model,
	duration time.Duration,
	msg string,
) (tea.Model, tea.Cmd) {

	var result string
	if strings.Compare(msg, "") == 0 {
		result = "" // TODO: eventually add tips.
	} else {
		result = msg
	}

	rtn := TransitionModel{
		usc:        usc,
		nextScreen: nextScreen,
		duration:   duration,
		msg:        result,

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

	return rtn, rtn.Init()

}

type doneMsg struct{}

func (m TransitionModel) Init() tea.Cmd {

	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(
			m.duration,
			func(t time.Time) tea.Msg {
				return doneMsg{}
			},
		),
		m.usc.LastWindowSizeReplay(),
	)

}

func (m TransitionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.usc.SetWindowsSize(msg)
		m.waitStyle = m.waitStyle.
			Height(msg.Height).
			Width(msg.Width)

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case doneMsg:
		return m.nextScreen, m.nextScreen.Init()

	}

	return m, tea.Batch(cmds...)

}

func (m TransitionModel) View() string {

	return m.waitStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.spinner.View(),
			m.msg,
		),
	)

}
