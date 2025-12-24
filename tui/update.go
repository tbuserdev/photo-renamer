package tui

import (
	"photo-renamer/renamer"

	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
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
	return m.FilePicker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.State == PreviewView {
				m.State = InputSelectView
				// Reset filepicker if needed, or just return to it
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if m.State == PreviewView {
				m.State = RenamingView
				// Use the already loaded PreviewActions if available
				// For in-place rename, InputPath is both source and dest
				return m, startRenaming(m.InputPath, m.InputPath, m.PreviewActions)
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
		s.Header = tableHeaderStyle.Copy().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ghBg2).
			BorderBottom(true)
		s.Selected = tableSelectedStyle
		s.Cell = tableCellStyle
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

	if m.State == InputSelectView {
		var cmd tea.Cmd
		m.FilePicker, cmd = m.FilePicker.Update(msg)

		if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
			m.InputPath = path
			m.State = PreviewView
			return m, startPreview(m.InputPath, m.InputPath)
		}

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			m.InputPath = m.FilePicker.CurrentDirectory
			m.State = PreviewView
			return m, startPreview(m.InputPath, m.InputPath)
		}

		return m, cmd
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
