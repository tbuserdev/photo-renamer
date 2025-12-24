package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// User Colors
	ghBg     = lipgloss.Color("#0d1117")
	ghRed    = lipgloss.Color("#fa7970")
	ghBg2    = lipgloss.Color("#161b22")
	ghOrange = lipgloss.Color("#faa356")
	ghBg3    = lipgloss.Color("#21262d")
	ghGray   = lipgloss.Color("#89929b")
	ghGreen  = lipgloss.Color("#7ce38b")
	ghBlueL  = lipgloss.Color("#a2d2fb")
	ghText   = lipgloss.Color("#c6cdd5")
	ghBlueM  = lipgloss.Color("#77bdfb")
	ghWhite  = lipgloss.Color("#ecf2f8")
	ghPurple = lipgloss.Color("#cea5fb")

	// Styles
	mainContainer = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ghBg).
			Background(ghBlueM).
			Padding(0, 1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(ghBlueL).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(ghBlueM).
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
)

func (m Model) View() string {
	var body string
	var footer string
	var headerTitle string

	switch m.State {
	case InputSelectView:
		headerTitle = "SELECT INPUT FOLDER"
		body = labelStyle.Render("LOCATION: ") + pathStyle.Render(m.FilePicker.CurrentDirectory) + "\n\n"
		body += m.FilePicker.View()
		footer = fmt.Sprintf("%s  •  %s  •  %s",
			keyStyle.Render("↑ ↓ ← →")+" navigate",
			keyStyle.Render("ENTER")+" preview changes",
			keyStyle.Render("ESC")+" quit",
		)

	case PreviewView:
		headerTitle = "PREVIEW RENAME"
		body = labelStyle.Render("PROPOSED CHANGES:") + "\n"
		body += tableContainer.Render(m.Table.View())
		footer = fmt.Sprintf("%s  •  %s",
			keyStyle.Render("ENTER")+" confirm renaming",
			keyStyle.Render("ESC")+" back",
		)

	case RenamingView:
		headerTitle = "RENAMING IN PROGRESS"
		body = m.ProgressBar.View() + "\n\n"
		body += fmt.Sprintf("Processing... %s / %s",
			keyStyle.Render(fmt.Sprintf("%d", m.ProcessedFiles)),
			labelStyle.Render(fmt.Sprintf("%d", m.TotalFiles)),
		)
		footer = "Please wait until the process is complete..."

	case DoneView:
		headerTitle = "PROCESS FINISHED"
		body = m.ProgressBar.View() + "\n\n"
		body += successStyle.Render(fmt.Sprintf("✓ Successfully processed %d files.", m.ProcessedFiles))
		if m.Err != nil {
			body += "\n\n" + errorStyle.Render("ERROR: ") + m.Err.Error()
		}
		footer = keyStyle.Render("ENTER") + " or " + keyStyle.Render("ESC") + " to quit"
	}

	header := headerStyle.Render(" " + headerTitle + " ")
	content := header + "\n" + body
	if footer != "" {
		content += "\n" + footerStyle.Render(helpStyle.Render(footer))
	}

	return mainContainer.Render(content)
}
