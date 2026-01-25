package screens

import (
	"time"

	"brisca.sh/server/requests"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

type userContext struct {
	Session     ssh.Session
	Renderer    *lipgloss.Renderer
	SizeMsg     tea.WindowSizeMsg
	Username    string
	ReqHandler  requests.Handler
	RenderEmoji bool

	LastRegisteredAction time.Time
}

type UserScreenContext struct {
	ctx userContext
}

func NewUserScreenContext(sess ssh.Session, rh requests.Handler) UserScreenContext {

	rtn := UserScreenContext{
		ctx: userContext{
			Session:     sess,
			Renderer:    bubbletea.MakeRenderer(sess),
			ReqHandler:  rh,
			RenderEmoji: true,

			LastRegisteredAction: time.Now(),
		},
	}

	return rtn

}

func (m UserScreenContext) Init() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
	)
}

func (m UserScreenContext) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.ctx.SizeMsg = msg
	}

	return m, nil

}

func (m UserScreenContext) View() string {
	return "empty view"
}

func (m *UserScreenContext) ReqHandler() *requests.Handler {
	return &m.ctx.ReqHandler
}

func (m *UserScreenContext) Renderer() *lipgloss.Renderer {
	return m.ctx.Renderer
}

func (m *UserScreenContext) SetLastRegisteredAction() {
	m.ctx.LastRegisteredAction = time.Now()
}

func (m *UserScreenContext) IsIdle() bool {
	return time.Now().After(
		m.ctx.LastRegisteredAction.Add(time.Minute),
	)
}

func (m *UserScreenContext) LastWindowSizeReplay() tea.Cmd {
	return func() tea.Msg {
		return m.ctx.SizeMsg
	}
}

func (m *UserScreenContext) SetWindowsSize(msg tea.WindowSizeMsg) {
	m.ctx.SizeMsg = msg
}

func (m *UserScreenContext) Username() string {
	return m.ctx.Username
}

func (m *UserScreenContext) SetUsername(name string) {
	m.ctx.Username = name
}

func (m *UserScreenContext) RenderEmoji() bool {
	return m.ctx.RenderEmoji
}

func (m *UserScreenContext) SetRenderEmoji(render bool) {
	m.ctx.RenderEmoji = render
}
