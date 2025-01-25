package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

var (
	upStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	downStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(75).
			Height(5).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	registerStyle = lipgloss.NewStyle().
			Width(75).
			Height(15).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69"))

	username string
)

type mainModel struct {
	state     sessionState
	textInput textinput.Model
	help      helpModel
	isUp      bool
	lobby     lobbyModel
}

func newModel() mainModel {
	m := mainModel{state: textInputView}
	m.textInput = textinput.New()
	m.textInput.Placeholder = "Guest"
	m.textInput.Focus()
	m.textInput.CharLimit = 25
	m.textInput.Width = 20
	m.textInput.Prompt = "\tWhat's your username?\n\t\t> "
	m.help = newHelp()
	m.isUp = statusRequest()
	m.lobby = newLobby()

	return m
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(textinput.Blink, m.help.Init(),
		tea.SetWindowTitle("brisca.sh"))
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		// h, v := docStyle.GetFrameSize()
		// m.lobby.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.help.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.help.keys.Help):
			m.help, _ = m.help.Update(msg)
			return m, nil
		case key.Matches(msg, m.help.keys.Enter):
			var register register
			register.Username = m.textInput.Value()
			if registerRequest(register) {
				username = register.Username
				m.lobby.list.Title = "User: " + username
				cmds = append(cmds, m.lobby.Init())
				lm, cmd := m.lobby.Update(msg)
				cmds = append(cmds, cmd)
				return lm, tea.Batch(cmds...)
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

func (m mainModel) View() string {
	switch {
	case m.state == textInputView:
		return registerView(m)
	}

	return ""
}

func registerView(m mainModel) string {
	var s string
	inside := "\t\t\t\t\t\t\tWelcome to brisca.sh!\n\n\n\n"
	inside += "Brisca Server: "
	if m.isUp {
		inside += upStyle.Render("Up")
	} else {
		inside += downStyle.Render("Down")
	}
	if username != "" {
		inside += "\tusername: " + username
	}
	inside += "\n\n" + m.textInput.View()
	s += lipgloss.JoinHorizontal(lipgloss.Top,
		registerStyle.Render(inside))
	s += helpStyle.Render(m.help.View())
	return s
}

func main() {

	logFile := "debug.log"
	envLogFile := os.Getenv("LOG_FILE")

	if envLogFile != "" {
		logFile = envLogFile
	}

	f, err := tea.LogToFile(logFile, "help")
	if err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	}
	log.SetOutput(f)
	defer f.Close() // nolint:errcheck

	if os.Getenv("HELP_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
		log.Helper()
		log.SetReportCaller(true)
		log.Debug("Debug Started")
	}

	log.Info("Current log Level ", log.GetLevel())

	p := tea.NewProgram(newModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	leaveGameRequest()
}
