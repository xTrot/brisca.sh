package screens

import (
	"math/rand"

	"brisca.sh/server/embedded"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	SIGNATURE = "  - brisca.sh"
)

type Model struct {
	suitArt string
	end     string
}

func NewModel(end string) (tea.Model, tea.Cmd) {

	pick := rand.Intn(4)
	var pickedSuit string

	switch pick {

	case 0:
		pickedSuit = embedded.BASTO

	case 1:
		pickedSuit = embedded.COPA

	case 2:
		pickedSuit = embedded.ESPADA

	case 3:
		pickedSuit = embedded.ORO
	default:
		pickedSuit = "Nothing picked error."

	}

	rtn := Model{
		suitArt: pickedSuit,
		end:     end,
	}

	return rtn, rtn.Init()
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Sequence(
		tea.ExitAltScreen,
		tea.ClearScreen,
		tea.Printf("%s\n\n%s\n%s\n",
			m.suitArt,
			m.end,
			SIGNATURE,
		),
		tea.Quit,
	)
}

// Update implements tea.Model.
func (m Model) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	return ""
}
