package tui

import (
	"ImageRenamer/renamer"

	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
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

	pb := progress.New(progress.WithDefaultGradient())

	return Model{
		State:       InputSelectView,
		FilePicker:  fp,
		ProgressBar: pb,
	}
}
