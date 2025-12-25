package tui

import (
	"fmt"
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
		footer = fmt.Sprintf("%s  •  %s  •  %s  •  %s",
			keyStyle.Render("↑/↓/←/→")+" navigate",
			keyStyle.Render("ENTER")+" show exif data",
			keyStyle.Render("R")+" preview rename",
			keyStyle.Render("ESC")+" quit",
		)

	case LoadingView:
		headerTitle = "SCANNING PREVIEW"
		body = fmt.Sprintf("\n %s Generating preview...", m.Spinner.View())
		footer = ""

	case PreviewView:
		headerTitle = "PREVIEW RENAME"
		stats := fmt.Sprintf("Total Files: %d  •  Original: %d", m.TotalFiles, m.OriginalFiles)
		body = labelStyle.Render("PROPOSED CHANGES:") + " " + helpStyle.Render(stats) + "\n"
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

	case DebugView:
		headerTitle = "FILE EXIF DEBUG"
		body = labelStyle.Render("EXIF DATA:") + "\n"
		body += tableContainer.Render(m.DebugTable.View())
		footer = keyStyle.Render("ESC") + " back"
	}

	header := headerStyle.Render(" " + headerTitle + " ")
	content := header + "\n" + body
	if footer != "" {
		content += "\n" + footerStyle.Render(helpStyle.Render(footer))
	}

	return mainContainer.Render(content)
}
