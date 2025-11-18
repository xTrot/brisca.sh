package screens

import (
	"time"

	"brisca.sh/server/requests"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type UserGlobal struct {
	Session     ssh.Session
	Renderer    *lipgloss.Renderer
	SizeMsg     tea.WindowSizeMsg
	Username    string
	ReqHandler  requests.Handler
	RenderEmoji bool

	LastRegisteredAction time.Time
}

func (m UserGlobal) LastWindowSizeReplay() tea.Cmd {
	return func() tea.Msg {
		return m.SizeMsg
	}
}

func (m UserGlobal) RegisterAction() tea.Cmd {
	return func() tea.Msg {
		m.LastRegisteredAction = time.Now()
		return nil
	}
}

func (m UserGlobal) IsIdle() bool {
	return time.Now().After(
		m.LastRegisteredAction.Add(time.Minute),
	)
}
