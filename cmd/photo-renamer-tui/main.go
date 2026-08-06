package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"photo-renamer/tui"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("photo-renamer-tui version: %s\n", Version)
		return
	}

	program := tea.NewProgram(tui.InitialModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "photo-renamer: %v\n", err)
		os.Exit(1)
	}
}
