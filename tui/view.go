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
		body = m.Styles.Label.Render("LOCATION: ") + m.Styles.Path.Render(m.FilePicker.CurrentDirectory) + "\n\n"
		body += m.FilePicker.View()
		footer = fmt.Sprintf("%s  •  %s  •  %s  •  %s  •  %s",
			m.Styles.Key.Render("↑/↓/←/→")+" navigate",
			m.Styles.Key.Render("ENTER")+" show exif data",
			m.Styles.Key.Render("R")+" preview rename",
			m.Styles.Key.Render("T")+" toggle theme",
			m.Styles.Key.Render("ESC")+" quit",
		)

	case LoadingView:
		headerTitle = "SCANNING PREVIEW"
		body = fmt.Sprintf("\n %s Generating preview...", m.Spinner.View())
		footer = ""

	case PreviewView:
		headerTitle = "PREVIEW RENAME"
		stats := fmt.Sprintf("Total Files: %d  •  Original: %d", m.TotalFiles, m.OriginalFiles)
		body = m.Styles.Label.Render("PROPOSED CHANGES:") + " " + m.Styles.Help.Render(stats) + "\n"
		body += m.Styles.TableContainer.Render(m.Table.View())
		footer = fmt.Sprintf("%s  •  %s",
			m.Styles.Key.Render("ENTER")+" confirm renaming",
			m.Styles.Key.Render("ESC")+" back",
		)

	case RenamingView:
		headerTitle = "RENAMING IN PROGRESS"
		body = m.ProgressBar.View() + "\n\n"
		body += fmt.Sprintf("Processing... %s / %s",
			m.Styles.Key.Render(fmt.Sprintf("%d", m.ProcessedFiles)),
			m.Styles.Label.Render(fmt.Sprintf("%d", m.TotalFiles)),
		)
		footer = "Please wait until the process is complete..."

	case DoneView:
		headerTitle = "PROCESS FINISHED"
		body = m.ProgressBar.View() + "\n\n"
		body += m.Styles.Success.Render(fmt.Sprintf("✓ Successfully processed %d files.", m.ProcessedFiles))
		if m.Err != nil {
			body += "\n\n" + m.Styles.Error.Render("ERROR: ") + m.Err.Error()
		}
		footer = m.Styles.Key.Render("ENTER") + " or " + m.Styles.Key.Render("ESC") + " to quit"

	case DebugView:
		headerTitle = "FILE EXIF DEBUG"
		body = m.Styles.Label.Render("EXIF DATA:") + "\n"
		body += m.Styles.TableContainer.Render(m.DebugTable.View())
		footer = m.Styles.Key.Render("ESC") + " back"
	}

	header := m.Styles.Header.Render(" " + headerTitle + " ")
	content := header + "\n" + body
	if footer != "" {
		content += "\n" + m.Styles.Footer.Render(m.Styles.Help.Render(footer))
	}

	return m.Styles.MainContainer.Render(content)
}
