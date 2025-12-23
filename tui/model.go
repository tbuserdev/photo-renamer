package tui

import (
	"ImageRenamer/renamer"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

type ValidState int

const (
	InputView ValidState = iota
	PreviewView
	RenamingView
	DoneView
)

type Model struct {
	State           ValidState
	InputPathInput  textinput.Model
	OutputPathInput textinput.Model
	Table           table.Model
	PreviewActions  []renamer.FileAction
	ProgressBar     progress.Model
	Progress        float64
	TotalFiles      int
	ProcessedFiles  int
	Err             error
	CurrentInput    int // 0 for InputPath, 1 for OutputPath
}

func InitialModel() Model {
	ipt := textinput.New()
	ipt.Placeholder = "Input Folder Path"
	ipt.Focus()
	ipt.Width = 50

	opt := textinput.New()
	opt.Placeholder = "Output Folder Path"
	opt.Width = 50

	pb := progress.New(progress.WithDefaultGradient())

	return Model{
		State:           InputView,
		InputPathInput:  ipt,
		OutputPathInput: opt,
		ProgressBar:     pb,
		CurrentInput:    0,
	}
}
