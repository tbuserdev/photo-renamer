package tui

import "github.com/charmbracelet/lipgloss"

var (
	// User Colors
	ghBg     = lipgloss.Color("#0d1117")
	ghBg2    = lipgloss.Color("#161b22")
	ghBg3    = lipgloss.Color("#21262d")
	ghGray   = lipgloss.Color("#89929b")
	ghText   = lipgloss.Color("#c6cdd5")
	ghWhite  = lipgloss.Color("#ecf2f8")
	ghRed    = lipgloss.Color("#fa7970")
	ghOrange = lipgloss.Color("#faa356")
	ghGreen  = lipgloss.Color("#7ce38b")
	ghBlueL  = lipgloss.Color("#a2d2fb")
	ghBlueM  = lipgloss.Color("#77bdfb")
	ghPurple = lipgloss.Color("#cea5fb")

	// Styles
	mainContainer = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ghBg).
			Background(ghRed).
			Padding(0, 1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(ghText).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(ghText).
			Italic(true).
			Underline(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(ghGray)

	keyStyle = lipgloss.NewStyle().
			Foreground(ghOrange).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ghBg2).
			PaddingTop(1)

	tableContainer = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ghBg2).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(ghRed).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(ghGreen).
			Bold(true)

	// Table Styles
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(ghWhite).
				Bold(true)

	tableSelectedStyle = lipgloss.NewStyle().
				Foreground(ghWhite).
				Background(ghBg3).
				Bold(false)

	tableCellStyle = lipgloss.NewStyle().
			Foreground(ghText)
)
