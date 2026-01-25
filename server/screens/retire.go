package screens

import tea "github.com/charmbracelet/bubbletea"

var (
	Retired = false
)

func HasRetiredUser(usc *UserScreenContext) (tea.Model, tea.Cmd) {

	if Retired && usc.IsIdle() {
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
