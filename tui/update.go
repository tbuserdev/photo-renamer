package tui

import (
	"ImageRenamer/renamer"

	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type progressMsg float64
type doneMsg struct{}
type errMsg error
type previewLoadedMsg []renamer.FileAction

// Global channel to signal progress from the blocking renamer
var globalProgressChan chan struct{}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if msg.String() == "esc" && m.State == PreviewView {
				m.State = InputView
				return m, nil
			}
			return m, tea.Quit
		case "tab", "shift+tab":
			if m.State == InputView {
				m.CurrentInput = (m.CurrentInput + 1) % 2
				if m.CurrentInput == 0 {
					m.InputPathInput.Focus()
					m.OutputPathInput.Blur()
				} else {
					m.InputPathInput.Blur()
					m.OutputPathInput.Focus()
				}
				return m, textinput.Blink
			}
		case "enter":
			if m.State == InputView {
				if m.InputPathInput.Value() != "" && m.OutputPathInput.Value() != "" {
					// Start Preview instead of direct renaming
					return m, startPreview(m.InputPathInput.Value(), m.OutputPathInput.Value())
				}
			} else if m.State == PreviewView {
				m.State = RenamingView
				// Use the already loaded PreviewActions if available
				return m, startRenaming(m.InputPathInput.Value(), m.OutputPathInput.Value(), m.PreviewActions)
			} else if m.State == DoneView {
				return m, tea.Quit
			}
		}
	case progressMsg:
		m.ProcessedFiles++
		var pct float64
		if m.TotalFiles > 0 {
			pct = float64(m.ProcessedFiles) / float64(m.TotalFiles)
		}
		cmd = m.ProgressBar.SetPercent(pct)
		return m, tea.Batch(cmd, waitForSingleProgress(globalProgressChan))
	case progress.FrameMsg:
		progressModel, cmd := m.ProgressBar.Update(msg)
		m.ProgressBar = progressModel.(progress.Model)
		return m, cmd
	case previewLoadedMsg:
		m.PreviewActions = msg
		var rows []table.Row
		for _, action := range msg {
			status := "OK"
			if action.IsError {
				status = "ERROR"
			} else if action.IsDuplicate {
				status = "DUPLICATE"
			}
			rows = append(rows, table.Row{
				filepath.Base(action.OriginalPath),
				action.NewName,
				status,
			})
		}

		columns := []table.Column{
			{Title: "Original", Width: 30},
			{Title: "New Name", Width: 40},
			{Title: "Status", Width: 10},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m.Table = t
		m.State = PreviewView
		return m, nil
	case renameStartedMsg:
		m.TotalFiles = msg.count
		m.ProcessedFiles = 0
		globalProgressChan = make(chan struct{})
		// Start the actual processing in a goroutine
		go func() {
			renamer.Rename(msg.actions, msg.output, msg.output+"/ERROR-OUTPUT", msg.output+"/DUPLICATES", func() {
				// Non-blocking send or buffered channel would be safer, but blocking is fine if consumer is fast
				globalProgressChan <- struct{}{}
			})
			close(globalProgressChan)
		}()
		return m, waitForSingleProgress(globalProgressChan)
	case doneMsg:
		m.State = DoneView
		return m, nil
	case errMsg:
		m.Err = msg
		m.State = DoneView
		return m, nil
	}

	if m.State == InputView {
		var cmdI, cmdO tea.Cmd
		m.InputPathInput, cmdI = m.InputPathInput.Update(msg)
		m.OutputPathInput, cmdO = m.OutputPathInput.Update(msg)
		return m, tea.Batch(cmdI, cmdO)
	}

	if m.State == PreviewView {
		m.Table, cmd = m.Table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func waitForSingleProgress(c chan struct{}) tea.Cmd {
	return func() tea.Msg {
		_, ok := <-c
		if !ok {
			return doneMsg{}
		}
		return progressMsg(1)
	}
}

func startRenaming(input, output string, existingActions []renamer.FileAction) tea.Cmd {
	return func() tea.Msg {
		var actions []renamer.FileAction
		var err error

		if len(existingActions) > 0 {
			actions = existingActions
		} else {
			// Direct start without preview, scan now
			actions, err = renamer.ScanFiles(input)
			if err != nil {
				return errMsg(err)
			}
		}

		count := len(actions)
		return renameStartedMsg{count: count, actions: actions, input: input, output: output}
	}
}

func startPreview(input, output string) tea.Cmd {
	return func() tea.Msg {
		actions, err := renamer.PreviewRename(input, output)
		if err != nil {
			return errMsg(err)
		}
		return previewLoadedMsg(actions)
	}
}

type renameStartedMsg struct {
	count   int
	actions []renamer.FileAction
	input   string
	output  string
}
