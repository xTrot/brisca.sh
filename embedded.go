package main

import (
	_ "embed"
)

var (
	//go:embed embededFiles/cheatSheet.md
	CheatSheet string
)
