package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	exitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Height(1).Padding(0).Margin(0)

	viewPortTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type MarkdownModel struct {
	Text  string
	Style lipgloss.Style
	ready bool

	viewport viewport.Model
	Title    string

	// Refers to the view being simple and static vs needing a viewport
	static bool
}

func NewCheatSheetModel() MarkdownModel {
	return MarkdownModel{
		Text:   CheatSheet,
		Style:  style,
		static: true,
	}
}

func NewFullHelpModel() MarkdownModel {
	viewport := viewport.New(24, 80)

	return MarkdownModel{
		Text:     FullHelp,
		Style:    style,
		static:   false,
		viewport: viewport,
		Title:    "How to Play brisca.sh",
	}
}

// title argument is only for viewport
func NewMarkdownModel(text string, static bool, title string) MarkdownModel {
	var vp viewport.Model
	if !static {
		vp = viewport.New(24, 80)
	}

	return MarkdownModel{
		Text:     text,
		Style:    style,
		static:   false,
		Title:    title,
		viewport: vp,
	}
}

func (m MarkdownModel) Init() tea.Cmd {
	return nil
}

func (m MarkdownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		helpHeight := 1
		verticalMarginHeight := headerHeight + footerHeight + helpHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			os.Setenv("GLAMOUR_STYLE", "dracula")
			renderer, _ := glamour.NewTermRenderer(
				glamour.WithWordWrap(m.Style.GetWidth()),
				glamour.WithEnvironmentConfig(),
			)
			out, err := renderer.Render(
				m.Text)
			if err != nil {
				log.Fatal(err.Error())
				panic("Markdown render panic.")
			}
			m.viewport.SetContent(out)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m MarkdownModel) View() string {
	if m.static {
		glamour.WithWordWrap(m.Style.GetWidth())
		out, err := glamour.Render(m.Text, "dracula")
		if err != nil {
			log.Fatal(err.Error())
			panic("Markdown render panic.")
		}
		return out
	} else {
		return m.ViewViewPort()
	}
}

func (m MarkdownModel) ViewViewPort() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m MarkdownModel) headerView() string {
	extraLine := "──"
	title := titleStyle.Render(m.Title)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)-2))
	return lipgloss.JoinHorizontal(lipgloss.Center, extraLine, title, line)
}

func (m MarkdownModel) footerView() string {
	extraLine := "──"
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)-2))
	s := lipgloss.JoinHorizontal(lipgloss.Center, line, info, extraLine)
	return lipgloss.JoinVertical(lipgloss.Center, s, exitStyle.Render("H exit this Help"))
}
