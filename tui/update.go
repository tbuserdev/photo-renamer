package tui

import (
	"photo-renamer/renamer"
	"sort"

	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidwall/gjson"
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
			switch m.State {
			case PreviewView, DebugView:
				m.State = InputSelectView
				// Reset filepicker if needed, or just return to it
				return m, nil
			default:
				return m, tea.Quit
			}
		case "enter":
			switch m.State {
			case PreviewView:
				m.State = RenamingView
				// Use the already loaded PreviewActions if available
				// For in-place rename, InputPath is both source and dest
				return m, startRenaming(m.InputPath, m.InputPath, m.PreviewActions)
			case DoneView:
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
			} else if action.IsSkipped {
				status = "SKIPPED"
			} else if action.IsDuplicate {
				status = "DUPLICATE"
			}
			rows = append(rows, table.Row{
				status,
				filepath.Base(action.OriginalPath),
				action.NewName,
			})
		}

		columns := []table.Column{
			{Title: "Status", Width: 10},
			{Title: "Original", Width: 40},
			{Title: "New Name", Width: 40},
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
			// User selected a file, show debug info
			jsonStr := renamer.GetExifData(path)
			m.DebugData = jsonStr // Keep raw just in case, or remove if unused

			// Parse JSON and build table
			var rows []table.Row
			result := gjson.Parse(jsonStr)

			// We want to sort keys for consistent display
			var keys []string
			result.ForEach(func(key, value gjson.Result) bool {
				keys = append(keys, key.String())
				return true
			})
			sort.Strings(keys)

			for _, k := range keys {
				val := result.Get(k)
				rows = append(rows, table.Row{k, val.String()})
			}

			columns := []table.Column{
				{Title: "Field", Width: 30},
				{Title: "Value", Width: 50},
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(15),
			)

			s := table.DefaultStyles()
			s.Header = tableHeaderStyle.Copy().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ghBg2).
				BorderBottom(true)
			s.Selected = tableSelectedStyle
			s.Cell = tableCellStyle
			t.SetStyles(s)

			m.DebugTable = t
			m.State = DebugView
			return m, nil
		}

		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "r", "R":
				m.InputPath = m.FilePicker.CurrentDirectory
				m.State = LoadingView
				return m, tea.Batch(m.Spinner.Tick, startPreview(m.InputPath, m.InputPath))
			}
		}

		return m, cmd
	}

	if m.State == PreviewView {
		m.Table, cmd = m.Table.Update(msg)
		return m, cmd
	}

	if m.State == DebugView {
		m.DebugTable, cmd = m.DebugTable.Update(msg)
		return m, cmd
	}

	if m.State == LoadingView {
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
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
