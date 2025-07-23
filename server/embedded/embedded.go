package embedded

import (
	_ "embed"
)

var (

	//go:embed embeddedFiles/cheatSheet.md
	CheatSheet string

	//go:embed embeddedFiles/fullHelp.md
	FullHelp string

	//go:embed embeddedFiles/makeGameHelp.md
	MakeGameHelp string

	//go:embed embeddedFiles/BASTO.art
	BASTO string

	//go:embed embeddedFiles/ESPADA.art
	ESPADA string

	//go:embed embeddedFiles/COPA.art
	COPA string

	//go:embed embeddedFiles/ORO.art
	ORO string

	//
)
