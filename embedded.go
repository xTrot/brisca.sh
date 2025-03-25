package main

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
)
