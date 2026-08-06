package gui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativeDialog "github.com/sqweek/dialog"

	"photo-renamer/renamer"
)

type phase int

const (
	phaseChoose phase = iota
	phaseScanning
	phaseReview
	phaseRenaming
	phaseComplete
	phaseFailed
)

type controller struct {
	window    fyne.Window
	content   *fyne.Container
	version   string
	phase     phase
	folder    string
	actions   []renamer.FileAction
	processed int
	total     int
	err       error
	cancel    context.CancelFunc
	scanID    uint64
}

// Run starts the production desktop application and blocks until it closes.
func Run(version string) {
	a := app.NewWithID("dev.tbuser.photo-renamer")
	a.Settings().SetTheme(newAppTheme())
	w := a.NewWindow("Photo Renamer")
	w.Resize(fyne.NewSize(980, 680))
	w.SetPadded(false)

	c := &controller{
		window:  w,
		content: container.NewStack(),
		version: version,
		phase:   phaseChoose,
	}
	w.SetCloseIntercept(c.closeRequested)
	c.content.Objects = []fyne.CanvasObject{c.page()}
	w.SetContent(c.content)
	w.ShowAndRun()
}

func (c *controller) render() {
	c.content.Objects = []fyne.CanvasObject{c.page()}
	c.content.Refresh()
}

func (c *controller) page() fyne.CanvasObject {
	body := c.body()
	footer := widget.NewLabelWithStyle(c.statusText(), fyne.TextAlignCenter, fyne.TextStyle{})
	return container.NewBorder(nil, container.NewVBox(widget.NewSeparator(), container.NewCenter(footer)), c.stepRail(), nil, container.NewPadded(body))
}

func (c *controller) body() fyne.CanvasObject {
	switch c.phase {
	case phaseChoose:
		return c.chooseView()
	case phaseScanning:
		return c.scanningView()
	case phaseReview:
		return c.reviewView()
	case phaseRenaming:
		return c.renamingView()
	case phaseComplete:
		return c.completeView()
	default:
		return c.failedView()
	}
}

func (c *controller) currentStep() int {
	switch c.phase {
	case phaseChoose, phaseScanning:
		return 0
	case phaseReview:
		return 1
	default:
		return 2
	}
}

func (c *controller) stepRail() fyne.CanvasObject {
	current := c.currentStep()
	labels := []string{"Choose folder", "Review changes", "Rename photos"}
	items := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Photo Renamer", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	}
	for index, label := range labels {
		marker := "○"
		style := fyne.TextStyle{}
		if index < current || c.phase == phaseComplete {
			marker = "✓"
		}
		if index == current && c.phase != phaseComplete {
			marker = "●"
			style.Bold = true
		}
		items = append(items, widget.NewLabelWithStyle(fmt.Sprintf("%s  %d   %s", marker, index+1, label), fyne.TextAlignLeading, style))
	}
	items = append(items,
		layout.NewSpacer(),
		widget.NewLabel("Previewed before changes.\nDuplicates stay recoverable."),
		widget.NewLabel("Version "+c.version),
	)
	return container.NewGridWrap(fyne.NewSize(220, 620), container.NewPadded(container.NewVBox(items...)))
}

func pageHeader(title, subtitle string) fyne.CanvasObject {
	heading := widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{SizeName: theme.SizeNameHeadingText, TextStyle: fyne.TextStyle{Bold: true}},
		Text:  title,
	})
	detail := widget.NewLabel(subtitle)
	detail.Wrapping = fyne.TextWrapWord
	return container.NewVBox(heading, detail, widget.NewSeparator())
}

func primaryButton(label string, icon fyne.Resource, tapped func()) *widget.Button {
	button := widget.NewButtonWithIcon(label, icon, tapped)
	button.Importance = widget.HighImportance
	return button
}

func (c *controller) chooseView() fyne.CanvasObject {
	path := widget.NewLabel("No folder selected")
	path.Alignment = fyne.TextAlignCenter
	path.Wrapping = fyne.TextWrapBreak
	if c.folder != "" {
		path.SetText(c.folder)
	}

	choose := widget.NewButtonWithIcon("Choose Folder…", theme.FolderOpenIcon(), c.chooseFolder)
	scan := primaryButton("Review Photos", theme.NavigateNextIcon(), c.startScan)
	if c.folder == "" {
		scan.Disable()
	}

	selection := widget.NewCard("", "", container.NewVBox(
		container.NewCenter(widget.NewIcon(theme.FolderOpenIcon())),
		path,
		container.NewCenter(choose),
	))
	selection.Resize(fyne.NewSize(580, 240))

	return container.NewBorder(
		pageHeader("Choose a photo folder", "Select the folder whose photos you want to organise. Nothing changes until you review and confirm."),
		container.NewHBox(layout.NewSpacer(), scan), nil, nil,
		container.NewCenter(container.NewGridWrap(fyne.NewSize(600, 270), selection)),
	)
}

func (c *controller) chooseFolder() {
	startDir := c.folder
	go func() {
		builder := nativeDialog.Directory().Title("Choose a photo folder")
		if startDir != "" {
			builder = builder.SetStartDir(startDir)
		}
		folder, err := builder.Browse()
		if err != nil || folder == "" {
			return
		}
		fyne.Do(func() {
			c.folder = folder
			c.err = nil
			c.render()
		})
	}()
}

func (c *controller) startScan() {
	if c.folder == "" {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.scanID++
	scanID := c.scanID
	c.cancel = cancel
	c.phase = phaseScanning
	c.err = nil
	c.render()
	folder := c.folder
	go func() {
		actions, err := renamer.PreviewRenameContext(ctx, folder, folder)
		fyne.Do(func() {
			if scanID != c.scanID {
				return
			}
			if errors.Is(err, context.Canceled) {
				return
			}
			c.cancel = nil
			if err != nil {
				c.err = err
				c.phase = phaseFailed
				c.render()
				return
			}
			c.actions = actions
			c.phase = phaseReview
			c.render()
		})
	}()
}

func (c *controller) scanningView() fyne.CanvasObject {
	activity := widget.NewActivity()
	activity.Start()
	card := widget.NewCard("", "", container.NewVBox(
		container.NewCenter(activity),
		widget.NewLabelWithStyle("Reading photo metadata…", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(filepath.Base(c.folder), fyne.TextAlignCenter, fyne.TextStyle{}),
		widget.NewLabelWithStyle("This may take a moment for large folders.", fyne.TextAlignCenter, fyne.TextStyle{}),
	))
	cancel := widget.NewButton("Cancel", func() {
		if c.cancel != nil {
			c.cancel()
		}
		c.cancel = nil
		c.scanID++
		c.phase = phaseChoose
		c.render()
	})
	return container.NewBorder(
		pageHeader("Preparing your preview", "Photo Renamer is reading capture dates and camera details with ExifTool."),
		container.NewHBox(cancel, layout.NewSpacer()), nil, nil,
		container.NewCenter(container.NewGridWrap(fyne.NewSize(560, 230), card)),
	)
}

func (c *controller) reviewView() fyne.CanvasObject {
	counts := summarize(c.actions)
	table := actionTable(c.actions)
	back := widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), func() {
		c.phase = phaseChoose
		c.actions = nil
		c.render()
	})
	process := primaryButton(processLabel(counts), theme.NavigateNextIcon(), c.confirmRename)
	if processableCount(counts) == 0 {
		process.Disable()
	}
	summary := fmt.Sprintf("%d ready  •  %d duplicate  •  %d error  •  %d unchanged", counts.ready, counts.duplicates, counts.errors, counts.skipped)
	return container.NewBorder(
		container.NewVBox(
			pageHeader("Review proposed changes", c.folder),
			container.NewBorder(nil, nil, widget.NewLabel(summary), widget.NewLabel("Drag column dividers to resize")),
		),
		container.NewHBox(back, layout.NewSpacer(), process), nil, nil, table,
	)
}

type actionCounts struct {
	ready, duplicates, errors, skipped int
}

func summarize(actions []renamer.FileAction) actionCounts {
	var counts actionCounts
	for _, action := range actions {
		switch {
		case action.IsError:
			counts.errors++
		case action.IsDuplicate:
			counts.duplicates++
		case action.IsSkipped:
			counts.skipped++
		default:
			counts.ready++
		}
	}
	return counts
}

func processLabel(counts actionCounts) string {
	count := processableCount(counts)
	if count == 0 {
		return "Nothing to Change"
	}
	if count == 1 {
		return "Process 1 Photo"
	}
	return fmt.Sprintf("Process %d Photos", count)
}

func processableCount(counts actionCounts) int {
	return counts.ready + counts.duplicates + counts.errors
}

func actionTable(actions []renamer.FileAction) fyne.CanvasObject {
	headings := []string{"Status", "Original", "New name / destination"}
	table := widget.NewTable(
		func() (int, int) { return len(actions), 3 },
		func() fyne.CanvasObject {
			label := widget.NewLabel("Minimum width")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			label := object.(*widget.Label)
			action := actions[id.Row]
			status, destination := actionPresentation(action)
			label.SetText([]string{status, filepath.Base(action.OriginalPath), destination}[id.Col])
		},
	)
	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabelWithStyle("Header", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	table.UpdateHeader = func(id widget.TableCellID, object fyne.CanvasObject) {
		if id.Row == -1 && id.Col >= 0 && id.Col < len(headings) {
			object.(*widget.Label).SetText(headings[id.Col])
		}
	}
	table.SetColumnWidth(0, 105)
	table.SetColumnWidth(1, 230)
	table.SetColumnWidth(2, 390)
	return table
}

func actionPresentation(action renamer.FileAction) (string, string) {
	switch {
	case action.IsError:
		reason := action.Error
		if reason == "" {
			reason = "Metadata could not be read"
		}
		return "⚠ Error", reason + " → ERROR-OUTPUT"
	case action.IsDuplicate:
		return "⧉ Duplicate", action.NewName + " → DUPLICATES"
	case action.IsSkipped:
		return "✓ Unchanged", action.NewName
	default:
		return "● Ready", action.NewName
	}
}

func (c *controller) confirmRename() {
	counts := summarize(c.actions)
	folder := c.folder
	go func() {
		ok := nativeDialog.Message(
			"Process %d photos in:\n%s\n\n%d will be renamed, %d duplicates and %d errors will be moved to review folders.",
			processableCount(counts), folder, counts.ready, counts.duplicates, counts.errors,
		).Title("Confirm Photo Changes").YesNo()
		if ok {
			fyne.Do(c.startRename)
		}
	}()
}

func (c *controller) startRename() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.phase = phaseRenaming
	c.processed = 0
	c.total = processableCount(summarize(c.actions))
	c.err = nil
	c.render()
	actions := append([]renamer.FileAction(nil), c.actions...)
	folder := c.folder
	go func() {
		err := renamer.ApplyPreviewContext(
			ctx,
			actions,
			folder,
			filepath.Join(folder, "ERROR-OUTPUT"),
			filepath.Join(folder, "DUPLICATES"),
			func(completed, total int) {
				fyne.DoAndWait(func() {
					c.processed = completed
					c.total = total
					c.render()
				})
			},
		)
		fyne.Do(func() {
			c.cancel = nil
			if err != nil {
				c.err = err
				c.phase = phaseFailed
			} else {
				c.phase = phaseComplete
			}
			c.render()
		})
	}()
}

func (c *controller) renamingView() fyne.CanvasObject {
	progress := widget.NewProgressBar()
	if c.total > 0 {
		progress.SetValue(float64(c.processed) / float64(c.total))
	}
	card := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle("Organising your photos…", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		progress,
		widget.NewLabelWithStyle(fmt.Sprintf("%d of %d", c.processed, c.total), fyne.TextAlignCenter, fyne.TextStyle{}),
		widget.NewLabelWithStyle("Keep this window open until processing finishes.", fyne.TextAlignCenter, fyne.TextStyle{}),
	))
	return container.NewBorder(
		pageHeader("Applying changes", c.folder), nil, nil, nil,
		container.NewCenter(container.NewGridWrap(fyne.NewSize(600, 220), card)),
	)
}

func (c *controller) completeView() fyne.CanvasObject {
	counts := summarize(c.actions)
	summary := widget.NewCard("", "", container.NewVBox(
		container.NewCenter(widget.NewIcon(theme.ConfirmIcon())),
		widget.NewLabelWithStyle("Your photos are organised", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(fmt.Sprintf("%d renamed  •  %d duplicates moved  •  %d errors moved", counts.ready, counts.duplicates, counts.errors), fyne.TextAlignCenter, fyne.TextStyle{}),
		widget.NewLabelWithStyle(c.folder, fyne.TextAlignCenter, fyne.TextStyle{}),
	))
	open := primaryButton("Show in File Manager", theme.FolderOpenIcon(), func() {
		if err := renamer.OpenOutputFolder(c.folder); err != nil {
			c.err = err
			c.phase = phaseFailed
			c.render()
		}
	})
	return container.NewBorder(
		pageHeader("Rename complete", "Every planned action finished successfully."),
		container.NewHBox(widget.NewButton("Rename Another Folder", c.reset), layout.NewSpacer(), open), nil, nil,
		container.NewCenter(container.NewGridWrap(fyne.NewSize(620, 240), summary)),
	)
}

func (c *controller) failedView() fyne.CanvasObject {
	message := "Something went wrong."
	if c.err != nil {
		message = c.err.Error()
	}
	message = strings.TrimSpace(message)
	detail := widget.NewLabel(message)
	detail.Wrapping = fyne.TextWrapWord
	subtitle := "No changes were made."
	if c.total > 0 {
		remaining := c.total - c.processed
		if remaining < 0 {
			remaining = 0
		}
		subtitle = fmt.Sprintf("%d of %d planned actions completed; %d remain. Review the selected folder before trying again.", c.processed, c.total, remaining)
	}
	card := widget.NewCard("Couldn’t complete this step", subtitle, detail)
	actions := container.NewHBox(widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), c.recoverFromFailure), layout.NewSpacer())
	if c.total > 0 {
		actions = container.NewHBox(
			widget.NewButton("Start Over", c.reset),
			layout.NewSpacer(),
			primaryButton("Show in File Manager", theme.FolderOpenIcon(), func() { _ = renamer.OpenOutputFolder(c.folder) }),
		)
	}
	return container.NewBorder(
		pageHeader("Photo Renamer needs your attention", "Review the details below before deciding what to do next."),
		actions, nil, nil,
		container.NewCenter(container.NewGridWrap(fyne.NewSize(620, 250), card)),
	)
}

func (c *controller) recoverFromFailure() {
	c.phase = phaseChoose
	c.actions = nil
	c.err = nil
	c.render()
}

func (c *controller) reset() {
	c.scanID++
	c.phase = phaseChoose
	c.folder = ""
	c.actions = nil
	c.processed = 0
	c.total = 0
	c.err = nil
	c.render()
}

func (c *controller) closeRequested() {
	if c.phase == phaseRenaming {
		go nativeDialog.Message("Photo Renamer is still processing files. Keep this window open until it finishes.").Title("Rename in Progress").Info()
		return
	}
	if c.cancel != nil {
		c.cancel()
	}
	c.window.SetCloseIntercept(nil)
	c.window.Close()
}

func (c *controller) statusText() string {
	switch c.phase {
	case phaseChoose:
		return "Choose one folder to begin"
	case phaseScanning:
		return "Reading metadata — no files are being changed"
	case phaseReview:
		return fmt.Sprintf("Preview ready — %d files found", len(c.actions))
	case phaseRenaming:
		return fmt.Sprintf("Processing %d of %d", c.processed, c.total)
	case phaseComplete:
		return "Complete"
	default:
		return "Action required"
	}
}
