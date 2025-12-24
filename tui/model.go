package tui

import (
	"ImageRenamer/renamer"

	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ValidState int

const (
	InputSelectView ValidState = iota
	PreviewView
	RenamingView
	DoneView
)

type Model struct {
	State          ValidState
	FilePicker     filepicker.Model
	InputPath      string
	Table          table.Model
	PreviewActions []renamer.FileAction
	ProgressBar    progress.Model
	Progress       float64
	TotalFiles     int
	ProcessedFiles int
	Err            error
}

func InitialModel() Model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	fp.CurrentDirectory, _ = os.Getwd()
	fp.AutoHeight = false
	fp.Height = 10
	fp.ShowPermissions = false
	fp.ShowSize = false

	// Apply Theme
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(ghOrange)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(ghBlueM)
	fp.Styles.File = lipgloss.NewStyle().Foreground(ghText)
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(ghPurple).Bold(true)
	fp.Styles.DisabledCursor = lipgloss.NewStyle().Foreground(ghGray)
	fp.Styles.EmptyDirectory = lipgloss.NewStyle().Foreground(ghGray).Italic(true)

	pb := progress.New(progress.WithGradient(string(ghBlueM), string(ghPurple)))

	return Model{
		State:       InputSelectView,
		FilePicker:  fp,
		ProgressBar: pb,
	}
}
