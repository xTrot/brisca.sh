package screens

import (
	"strings"
	"time"
	"unicode"

	"brisca.sh/server/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

// sessionState is used to track which model is focused
type sessionState uint

const (
	defaultTime                = time.Minute
	textInputView sessionState = iota
)

type RegisterModel struct {
	state      sessionState
	textInput  textinput.Model
	help       HelpModel
	isUp       bool
	userGlobal UserGlobal
	browser    string

	upStyle       lipgloss.Style
	downStyle     lipgloss.Style
	helpStyle     lipgloss.Style
	registerStyle lipgloss.Style
}

func NewRegisterScreen(session *ssh.Session, browser string, gameServer string) RegisterModel {

	m := RegisterModel{
		state:   textInputView,
		browser: browser,
	}
	m.textInput = textinput.New()
	m.textInput.Placeholder = "Guest"
	m.textInput.Focus()
	m.textInput.CharLimit = 25
	m.textInput.Width = 20
	m.textInput.Prompt = "\tWhat's your username?\n\t\t> "
	m.help = NewHelp()
	m.userGlobal = UserGlobal{
		Session:     *session,
		Renderer:    bubbletea.MakeRenderer(*session),
		ReqHandler:  requests.NewHandler(browser, gameServer),
		RenderEmoji: true,

		LastRegisteredAction: time.Now(),
	}
	m.isUp = m.userGlobal.ReqHandler.StatusRequest(requests.BROWSER)

	m.upStyle = m.userGlobal.Renderer.NewStyle().Foreground(lipgloss.Color("10"))
	m.downStyle = m.userGlobal.Renderer.NewStyle().Foreground(lipgloss.Color("9"))
	m.helpStyle = m.userGlobal.Renderer.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(75).Height(5).
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder())
	m.registerStyle = m.userGlobal.Renderer.NewStyle().
		Width(75).Height(15).
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("69"))

	return m
}

func (m RegisterModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(textinput.Blink, m.help.Init(),
		tea.SetWindowTitle("brisca.sh"))
}

func (m RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	quit, cmd := HasRetiredUser(m.userGlobal)
	if cmd != nil {
		return quit, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.userGlobal.SizeMsg = msg

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.help.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.help.Keys.Help):
			m.help, _ = m.help.Update(msg)
			return m, nil
		case key.Matches(msg, m.help.Keys.Enter):
			var register requests.Register
			register.Username = m.textInput.Value()
			if m.userGlobal.ReqHandler.RegisterRequest(register, m.browser) {
				m.userGlobal.Username = register.Username
				lm := NewLobby(m.userGlobal)
				return lm, tea.Batch(lm.Init())
			}
		}
		charRune := []rune(msg.String())[0]
		if m.state == textInputView && okChars(charRune) {
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		}

	}
	return m, tea.Batch(cmds...)
}

func okChars(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if unicode.IsDigit(r) {
		return true
	}
	if strings.ContainsRune("@-.+_?", r) { // ok Symbols
		return true
	}
	return false
}

func (m RegisterModel) View() string {
	switch m.state {
	case textInputView:
		return registerView(m)
	}

	return ""
}

func registerView(m RegisterModel) string {
	var s string
	inside := "\t\t\t\t\t\t\tWelcome to brisca.sh!\n\n\n\n"
	inside += "Brisca Server: "
	if m.isUp {
		inside += m.upStyle.Render("Up")
	} else {
		inside += m.downStyle.Render("Down")
	}
	inside += "\n\n" + m.textInput.View()
	s += lipgloss.JoinHorizontal(lipgloss.Top,
		m.registerStyle.Render(inside))
	s += m.helpStyle.Render(m.help.View())
	return s
}
