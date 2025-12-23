package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			MarginBottom(1)

	focusedStyle = inputStyle.Copy().
			BorderForeground(lipgloss.Color("205"))

	btnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginTop(1)
)

func (m Model) View() string {
	var s string

	switch m.State {
	case InputView:
		s = titleStyle.Render("Image Rename & Copy Tool") + "\n\n"

		s += "Input Folder:\n"
		if m.CurrentInput == 0 {
			s += focusedStyle.Render(m.InputPathInput.View()) + "\n"
		} else {
			s += inputStyle.Render(m.InputPathInput.View()) + "\n"
		}

		s += "Output Folder:\n"
		if m.CurrentInput == 1 {
			s += focusedStyle.Render(m.OutputPathInput.View()) + "\n"
		} else {
			s += inputStyle.Render(m.OutputPathInput.View()) + "\n"
		}

		s += "\n[TAB] Switch | [ENTER] Start | [ESC] Quit\n"

	case RenamingView:
		s = titleStyle.Render("Renaming...") + "\n\n"
		s += m.ProgressBar.View() + "\n\n"
		s += fmt.Sprintf("Processing... %d / %d", m.ProcessedFiles, m.TotalFiles) + "\n"

	case DoneView:
		s = titleStyle.Render("Process Finished") + "\n\n"
		s += m.ProgressBar.View() + "\n\n"
		s += fmt.Sprintf("Completed: %d files renamed.", m.ProcessedFiles) + "\n"
		if m.Err != nil {
			s += fmt.Sprintf("\nError: %v\n", m.Err)
		}
		s += "\n[ENTER] Quit\n"
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(s)
}
