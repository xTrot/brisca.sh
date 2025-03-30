package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"golang.org/x/exp/rand"
)

var (
	scoreCounterStyle = lipgloss.NewStyle()
)

type scoreCounter struct {
	index        int
	name         string
	cards        []card
	style        lipgloss.Style
	spinner      spinner.Model
	countedCards []countedCard
	counted      int
	total        int

	renderEmoji bool
}

type countedCard struct {
	duration time.Duration
	card     card
	total    int
}

func newScoreCounter(index int, name string, cards []card, renderEmoji bool) scoreCounter {
	const showLastResults = 5

	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("206"))

	return scoreCounter{
		index: index,
		name:  name,
		cards: cards,
		style: lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(1),
		spinner:      sp,
		countedCards: make([]countedCard, showLastResults),
		renderEmoji:  renderEmoji,
	}
}

func (m scoreCounter) Init() tea.Cmd {
	if len(m.cards) > 0 {
		return tea.Batch(
			m.spinner.Tick,
			m.runPretendCount(m.cards[m.counted]),
		)
	} else {
		return tea.Batch(
			m.spinner.Tick,
			m.noCards(),
		)
	}
}

func (m scoreCounter) noCards() tea.Cmd {
	return func() tea.Msg {
		return doneCounting{m.index}
	}
}

type stopSpinnerMsg tea.Msg

func (m scoreCounter) Update(msg tea.Msg) (scoreCounter, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		if len(m.cards) == m.counted {
			return m, nil
		}
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case pretendCountMsg:
		d := time.Duration(msg.time)
		m.total += msg.card.score
		res := countedCard{card: msg.card, duration: d, total: m.total}
		m.countedCards = append(m.countedCards[1:], res)
		m.counted++
		if len(m.cards) == m.counted {
			return m, func() tea.Msg {
				return doneCounting{index: m.index}
			}
		}
		return m, m.runPretendCount(m.cards[m.counted])
	default:
		return m, nil
	}
}

func (m scoreCounter) View() string {
	s := m.name + ":\n\n"
	if len(m.cards) != m.counted {
		s += m.spinner.View() + " Counting cards..."
	} else {
		s += fmt.Sprintf("Total:%3d", m.total)
	}
	s += "\n\n"

	for _, res := range m.countedCards {
		if res.duration == 0 {
			s += "..........................\n" // Width 26 equal to else statement
		} else {
			s += fmt.Sprintf("%s Worth:%2d Tally:%3d\n", // Width 26 equal to if statement
				res.card.renderCard(m.renderEmoji), res.card.score, res.total)
		}
	}

	return m.style.Render(s)
}

// pretendCountMsg is sent when a pretend process completes.
type pretendCountMsg struct {
	time time.Duration
	card card
	id   int
}

// pretendProcess simulates a long-running process.
func (m scoreCounter) runPretendCount(card card) tea.Cmd {
	return func() tea.Msg {
		pause := time.Duration(rand.Int63n(499)+100) * time.Millisecond
		time.Sleep(pause)
		return pretendCountMsg{time: pause, card: card, id: m.index}
	}
}
