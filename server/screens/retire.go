package screens

import tea "github.com/charmbracelet/bubbletea"

var (
	Retired = false
)

func HasRetiredUser(ug UserGlobal) (tea.Model, tea.Cmd) {

	if Retired && ug.IsIdle() {
		return NewModel("Your server has been retired, please reconnect.")
	} else {
		return nil, nil
	}

}

func HasRetired() (tea.Model, tea.Cmd) {

	if Retired {
		return NewModel("Your server has been retired, please reconnect.")
	} else {
		return nil, nil
	}

}
