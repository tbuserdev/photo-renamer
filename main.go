package main

import (
	"ImageRenamer/progressCounter"
	"ImageRenamer/renamer"
	"ImageRenamer/renamer/utility"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"time"
)

func main() {
	// Folders
	duplicateFolder := "DUPLICATES"
	errorFolder := "ERROR-OUTPUT"

	// Window
	myApp := app.New()
	window := myApp.NewWindow("Image Rename & Copy Tool")

	// Window Size
	standardWindow := fyne.NewSize(400, 4*50)
	FileDialogWindow := fyne.NewSize(600, 400)
	window.Resize(standardWindow)

	// Widgets
	// TEXT - "Input"
	title := widget.NewLabel(" -- Input -- ")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle.Bold = true
	title.TextStyle.Monospace = true

	// INPUT
	// ENTRY - Input File Path
	inputFilepath := widget.NewEntry()
	inputFilepath.SetPlaceHolder("Choose your Input Path (Camera)")

	// BUTTON - Input File Path with Dialog
	inputFolderButton := widget.NewButton("...", func() {
		window.Resize(FileDialogWindow)
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err == nil && list != nil {
				inputFilepath.SetText(list.Path())
			} else {
				inputFilepath.SetText("")
			}
			window.Resize(standardWindow)
		}, window)
	})

	// OUTPUT
	// ENTRY - Output File Path
	inputOutputPath := widget.NewEntry()
	inputOutputPath.SetPlaceHolder("Choose your Output Path (PC/NAS)")

	// BUTTON - Output File Path with Dialog
	outputFolderButton := widget.NewButton("...", func() {
		window.Resize(FileDialogWindow)
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err == nil && list != nil {
				inputOutputPath.SetText(list.Path())
			} else {
				inputOutputPath.SetText("")
			}
			window.Resize(standardWindow)
		}, window)
	})
	outputFolderButton.Resize(fyne.NewSize(50, 50))

	startButton := widget.NewButton("Start", func() {})
	resetButton := widget.NewButton("Reset", func() {})
	openOutputButton := widget.NewButton("Open Output", func() {})
	closeButton := widget.NewButton("Close", func() { myApp.Quit() })

	progressBar := widget.NewProgressBar()

	errorLabel := widget.NewLabel("ERROR: Please choose a valid input and output path!")
	errorLabel.Alignment = fyne.TextAlignCenter
	errorLabel.TextStyle.Bold = true

	// Components
	inputLayout := container.New(layout.NewFormLayout(), inputFolderButton, inputFilepath)
	outputLayout := container.New(layout.NewFormLayout(), outputFolderButton, inputOutputPath)

	// Layout
	startLayout := container.NewVBox(inputLayout, outputLayout, startButton, resetButton, errorLabel)
	progressLayout := container.NewVBox(progressBar, errorLabel, closeButton, openOutputButton)
	mainContent := container.NewVBox(title, startLayout, progressLayout)

	// Hide
	progressLayout.Hide()
	errorLabel.Hide()
	openOutputButton.Hide()
	closeButton.Hide()

	// Events
	startButton.OnTapped = func() {
		if inputFilepath.Text == "" || inputOutputPath.Text == "" {
			go func() {
				errorLabel.Show()
				time.Sleep(5 * time.Second)
				errorLabel.Hide()
			}()
		} else {
			startLayout.Hide()
			errorLabel.Hide()
			title.SetText(" -- Renaming... -- ")
			progressLayout.Show()

			count, err := progressCounter.CountFiles(inputFilepath.Text)
			if err != nil {
				errorLabel.Show()
				errorLabel.SetText("ERROR: " + err.Error())
			}
			progressBar.SetValue(0.0)
			progressBar.Max = float64(count)

			inputFolder := inputFilepath.Text
			outputFolder := inputOutputPath.Text
			func() {
				err := renamer.Rename(inputFolder, outputFolder, errorFolder, duplicateFolder, progressBar)
				if err != nil {
					errorLabel.Show()
					errorLabel.SetText("ERROR: " + err.Error())
				} else {
					title.SetText(" -- PROCESS FINISHED -- ")
					openOutputButton.Show()
					closeButton.Show()
				}
			}()
		}
	}

	resetButton.OnTapped = func() {
		inputFilepath.SetText("")
		inputOutputPath.SetText("")
		errorLabel.Hide()
	}

	openOutputButton.OnTapped = func() {
		err := utility.OpenOutputFolder(inputOutputPath.Text)
		if err != nil {
			errorLabel.Show()
			errorLabel.SetText("ERROR: " + err.Error())
		}
	}

	window.SetContent(mainContent)
	window.ShowAndRun()
}
