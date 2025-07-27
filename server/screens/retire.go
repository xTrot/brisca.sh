package screens

import tea "github.com/charmbracelet/bubbletea"

var (
	Retired = false
)

func HasRetired() (tea.Model, tea.Cmd) {

	if Retired {
		return NewModel("Your server has been retired, please reconnect.")
	} else {
		return nil, nil
	}

}
