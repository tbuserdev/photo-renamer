package tui

import (
	"ImageRenamer/progressCounter"
	"ImageRenamer/renamer"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type progressMsg float64
type doneMsg struct{}
type errMsg error

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
					m.State = RenamingView
					return m, startRenaming(m.InputPathInput.Value(), m.OutputPathInput.Value())
				}
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
	case renameStartedMsg:
		m.TotalFiles = msg.count
		m.ProcessedFiles = 0
		globalProgressChan = make(chan struct{})
		// Start the actual processing in a goroutine
		go func() {
			renamer.Rename(msg.input, msg.output, msg.output+"/ERROR-OUTPUT", msg.output+"/DUPLICATES", func() {
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

func startRenaming(input, output string) tea.Cmd {
	return func() tea.Msg {
		// Count files
		count, err := progressCounter.CountFiles(input)
		if err != nil {
			return errMsg(err)
		}
		return renameStartedMsg{count: count, input: input, output: output}
	}
}

type renameStartedMsg struct {
	count  int
	input  string
	output string
}
