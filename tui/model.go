package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
)

type ValidState int

const (
	InputView ValidState = iota
	RenamingView
	DoneView
)

type Model struct {
	State           ValidState
	InputPathInput  textinput.Model
	OutputPathInput textinput.Model
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
