package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	emptyBoxStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	tableBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("82"))
	playerBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	activePlayerBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	gsHelpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("241"))
	selectedCardStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("69"))
	windowWidthMin  = 80
	windowHighttMin = 24
)

type box struct {
	view  string
	style lipgloss.Style
}

type gsModel struct {
	timer        timer.Model
	spinner      spinner.Model
	index        int
	boxes        [3][3]box
	hand         []card
	selectedCard int
}

func newGSModel(timeout time.Duration) gsModel {
	m := gsModel{}
	m.timer = timer.New(timeout)
	m.spinner = spinner.New()
	var boxes [3][3]box
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			boxes[i][j] = box{
				view: " ",
			}
		}
	}
	m.boxes = boxes
	m.selectedCard = 0
	return m
}

func (m gsModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.timer.Init(), m.spinner.Tick, tea.WindowSize())
}

func (m gsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case resizeMsg:
		return msg.model, nil
	case tea.WindowSizeMsg:
		return m.updateWindow(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			handSize := len(m.hand)
			rawMove := m.selectedCard - 1
			m.selectedCard = (rawMove%handSize + handSize) % handSize
			return m, nil
		case "right":
			m.selectedCard = (m.selectedCard + 1) % len(m.hand)
			return m, nil
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd, updateHand)
	case updateHandMsg:
		m.hand = msg.hand
	}
	return m, tea.Batch(cmds...)
}

type resizeMsg struct {
	model gsModel
}

func (m gsModel) updateWindow(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		wholeWidth := max(msg.Width, windowWidthMin) - 6        // 6 to account for borders
		wholeHeight := max(msg.Height, windowHighttMin) - 6 - 3 // Space reserved for bars
		thirdWidth := wholeWidth / 3
		thirdHeight := wholeHeight / 3
		midWidth := wholeWidth - (2 * thirdWidth)
		midHeight := wholeHeight - (2 * thirdHeight)

		// Blank Top Left
		m.boxes[0][0].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		// Second Player(2,3), Third Player(4)
		m.boxes[0][1].style = playerBoxStyle.Width(midWidth).Height(thirdHeight)

		// Blank Top Right
		m.boxes[0][2].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)
		// Second Player(4)
		m.boxes[1][0].style = emptyBoxStyle.Width(thirdWidth).Height(midHeight)

		// Table
		m.boxes[1][1].style = tableBoxStyle.Width(midWidth).Height(midHeight)

		// Last Player(3,4)
		m.boxes[1][2].style = playerBoxStyle.Width(thirdWidth).Height(midHeight)

		// Blank Bottom Left
		m.boxes[2][0].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		// This player
		m.boxes[2][1].style = activePlayerBoxStyle.Width(midWidth).Height(thirdHeight)

		// Blank Bottom Right
		m.boxes[2][2].style = emptyBoxStyle.Width(thirdWidth).Height(thirdHeight)

		return resizeMsg{model: m}
	}
}

func (m gsModel) View() string {
	var s string
	m.boxes[0][1].view = "Player2 Score: 10\n  Score Pile:\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7] ..."
	// m.boxes[0][2].view =
	// m.boxes[1][0].view =
	m.boxes[1][1].view = "Table:\n  Deck: 40\n  Suit Card: [🪵: 3]\n\n  In Play:\n  [⚔️: 1] [🪙: 5] [🏆:10]"
	m.boxes[1][2].view = "1234512345123451234512345 Score: 10\n  Score Pile:\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7] ..."
	// m.boxes[2][0].view =
	m.boxes[2][1].view = "xTrot Score: 10\n  Score Pile:\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7]\n  [⚔️: 3] [🪵: 2] [🏆: 6] [🏆: 7] ..."
	// m.boxes[2][2].view =
	for i := 0; i < len(m.boxes); i++ {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			m.boxes[i][0].style.Render(m.boxes[i][0].view),
			m.boxes[i][1].style.Render(m.boxes[i][1].view),
			m.boxes[i][2].style.Render(m.boxes[i][2].view),
		)
		s = lipgloss.JoinVertical(lipgloss.Top, s, row)
	}
	s = lipgloss.JoinVertical(lipgloss.Top, s, m.handView())
	s = lipgloss.JoinVertical(lipgloss.Top, s, lipgloss.JoinHorizontal(lipgloss.Left, "Status: Your Turn, timer: ", m.timer.View()))
	s = lipgloss.JoinVertical(lipgloss.Center, s, gsHelpStyle.Render("← →: select card      enter: play card"))
	return s
}

func (m *gsModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *gsModel) handView() string {
	var s string
	s = "Hand:"
	for i := range len(m.hand) {
		card := m.hand[i]
		if m.selectedCard == i {
			s += fmt.Sprintf("%2d:%s", i+1, selectedCardStyle.Render(renderCard(card)))
		} else {
			s += fmt.Sprintf("%2d:%s", i+1, renderCard(card))
		}
	}
	return s
}

func renderCard(card card) string {
	return fmt.Sprintf("[%s:%2d]", card.suit, card.num)
}

type updateHandMsg struct {
	hand []card
}

func updateHand() tea.Msg {
	newHand := handRequest()

	return updateHandMsg{
		hand: newHand,
	}
}
