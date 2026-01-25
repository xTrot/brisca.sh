package screens

import (
	"fmt"
	"time"

	"brisca.sh/server/game"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	DEBOUNCE_TIME = time.Second
)

type winScreen struct {
	scArray    [3]scoreCounter
	scSize     int
	countDone  int
	gameConfig game.GameConfigPayload
	winString  string
	usc        UserScreenContext
	debounced  bool

	style       lipgloss.Style
	winnerStyle lipgloss.Style
	helpStyle   lipgloss.Style
}

type debounceMsg struct{}

func debounce() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(DEBOUNCE_TIME)
		return debounceMsg{}
	}
}

type doneCounting struct {
	index int
}

func newWinScreen(
	gc *game.GameConfigPayload,
	players []playerModel,
	gameWon *game.GameWonPayload,
	usc UserScreenContext,
) winScreen {

	var firstScoreCounter scoreCounter
	var secondScoreCounter scoreCounter
	var thirdScoreCounter scoreCounter
	var scSize int
	var winString string

	switch gc.MaxPlayers {
	case 2:
		firstScoreCounter = newScoreCounter(0, players[0].name, players[0].scorePile, usc.RenderEmoji(), usc.Renderer())
		secondScoreCounter = newScoreCounter(1, players[1].name, players[1].scorePile, usc.RenderEmoji(), usc.Renderer())
		scSize = 2
		switch gameWon.Seat {
		case -1:
			winString = "It was a tie!"
		default:
			winString = players[gameWon.Seat].name + " won!!!"
		}
	case 3:
		firstScoreCounter = newScoreCounter(0, players[0].name, players[0].scorePile, usc.RenderEmoji(), usc.Renderer())
		secondScoreCounter = newScoreCounter(1, players[1].name, players[1].scorePile, usc.RenderEmoji(), usc.Renderer())
		thirdScoreCounter = newScoreCounter(2, players[2].name, players[2].scorePile, usc.RenderEmoji(), usc.Renderer())
		scSize = 3
		switch gameWon.Seat {
		case -1:
			winString = "It was a tie!"
		default:
			winString = players[gameWon.Seat].name + " won!!!"
		}
	case 4:
		teamAString := fmt.Sprintf("Team A:\n %s and %s", players[0].name, players[2].name)
		teamBString := fmt.Sprintf("Team B:\n %s and %s", players[1].name, players[3].name)
		firstScoreCounter = newScoreCounter(0, "Team A", append(players[0].scorePile, players[2].scorePile...), usc.RenderEmoji(), usc.Renderer())
		secondScoreCounter = newScoreCounter(1, "Team B", append(players[1].scorePile, players[3].scorePile...), usc.RenderEmoji(), usc.Renderer())
		scSize = 2
		switch gameWon.Team {
		case "A":
			winString = teamAString + " won!!!"
		case "B":
			winString = teamBString + " won!!!"
		case "draw":
			winString = "It was a tie!"
		}
	default:
		panic(fmt.Sprintf("gameConfig.MaxPlayers not 2-4, gameConfig=%v", gc))
	}

	m := winScreen{
		gameConfig: *gc,
		scArray: [3]scoreCounter{
			firstScoreCounter,
			secondScoreCounter,
			thirdScoreCounter,
		},
		scSize:    scSize,
		winString: winString,
		usc:       usc,
	}

	m.style = m.usc.Renderer().NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("69"))
	m.winnerStyle = m.usc.Renderer().NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder()).
		BorderForeground(lipgloss.Color("69"))
	m.helpStyle = m.usc.Renderer().NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Left, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder())

	return m

}

func (m winScreen) Init() tea.Cmd {
	if m.gameConfig.MaxPlayers == 3 {
		return tea.Batch(
			m.usc.LastWindowSizeReplay(),
			m.scArray[0].Init(),
			m.scArray[1].Init(),
			m.scArray[2].Init(),
			debounce(),
		)
	} else {
		return tea.Batch(
			m.usc.LastWindowSizeReplay(),
			m.scArray[0].Init(),
			m.scArray[1].Init(),
			debounce(),
		)
	}
}

func (m winScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	quit, cmd := HasRetiredUser(&m.usc)
	if cmd != nil {
		return quit, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.usc.SetWindowsSize(msg)
		m.style = m.style.
			Width(max(windowWidthMin, msg.Width) - 2).
			Height(max(windowHighttMin, msg.Height) - 2)

	case debounceMsg:
		m.debounced = true

	case tea.KeyMsg:
		if m.debounced {
			quit, cmd = HasRetired()
			if cmd != nil {
				return quit, cmd
			}
			lm := NewLobby(m.usc)
			return lm, lm.Init()
		}
	case pretendCountMsg:
		switch msg.id {
		case 0:
			m.scArray[0], cmd = m.scArray[0].Update(msg)
			cmds = append(cmds, cmd)
		case 1:
			m.scArray[1], cmd = m.scArray[1].Update(msg)
			cmds = append(cmds, cmd)
		case 2:
			m.scArray[2], cmd = m.scArray[2].Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.scArray[0], cmd = m.scArray[0].Update(msg)
		cmds = append(cmds, cmd)
		m.scArray[1], cmd = m.scArray[1].Update(msg)
		cmds = append(cmds, cmd)
		m.scArray[2], cmd = m.scArray[2].Update(msg)
		cmds = append(cmds, cmd)
	case doneCounting:
		m.countDone++
		log.Debug("case doneCounting:", "msg", msg)
	}

	return m, tea.Batch(cmds...)
}

func (m winScreen) View() string {
	var s string
	if m.gameConfig.MaxPlayers == 3 {
		s = lipgloss.JoinHorizontal(lipgloss.Center,
			m.scArray[0].View(), m.scArray[1].View(), m.scArray[2].View())
	} else {
		s = lipgloss.JoinHorizontal(lipgloss.Center,
			m.scArray[0].View(), m.scArray[1].View())
	}

	winner := m.winString
	if m.countDone != m.scSize {
		winner = " "
	}

	s = lipgloss.JoinVertical(lipgloss.Center,
		m.winnerStyle.Render(winner),
		s,
		m.helpStyle.AlignHorizontal(lipgloss.Center).Render("Press any key to exit"))

	return m.style.Render(s)
}
