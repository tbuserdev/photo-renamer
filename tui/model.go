package tui

import (
	"photo-renamer/renamer"

	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type ValidState int

const (
	InputSelectView ValidState = iota
	LoadingView
	PreviewView
	RenamingView
	DoneView
	DebugView
)

type Model struct {
	State          ValidState
	FilePicker     filepicker.Model
	InputPath      string
	Spinner        spinner.Model
	Table          table.Model
	PreviewActions []renamer.FileAction
	ProgressBar    progress.Model
	Progress       float64
	TotalFiles     int
	ProcessedFiles int
	OriginalFiles  int
	Err            error
	DebugData      string
	DebugTable     table.Model
	Styles         Styles
	Theme          Theme
}

func InitialModel() Model {
	theme := FlexokiDark
	if !lipgloss.HasDarkBackground() {
		theme = FlexokiLight
	}
	styles := InitStyles(theme)

	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = true
	fp.ShowHidden = false
	fp.CurrentDirectory, _ = os.Getwd()
	fp.AutoHeight = false
	fp.Height = 10
	fp.ShowPermissions = false
	fp.ShowSize = false

	// Apply Theme
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(theme.Orange)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(theme.Blue)
	fp.Styles.File = lipgloss.NewStyle().Foreground(theme.Fg)
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(theme.Purple).Bold(true)
	fp.Styles.DisabledCursor = lipgloss.NewStyle().Foreground(theme.Muted)
	fp.Styles.EmptyDirectory = lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)

	pb := progress.New(progress.WithGradient(string(theme.Blue), string(theme.Purple)))

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Purple)

	return Model{
		State:       InputSelectView,
		FilePicker:  fp,
		ProgressBar: pb,
		Spinner:     s,
		Styles:      styles,
		Theme:       theme,
	}
}
