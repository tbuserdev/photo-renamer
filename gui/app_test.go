package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"photo-renamer/renamer"
)

func TestSummarize(t *testing.T) {
	got := summarize([]renamer.FileAction{
		{},
		{IsDuplicate: true},
		{IsError: true},
		{IsSkipped: true},
	})
	want := actionCounts{ready: 1, duplicates: 1, errors: 1, skipped: 1}
	if got != want {
		t.Fatalf("summarize() = %+v, want %+v", got, want)
	}
}

func TestActionPresentation(t *testing.T) {
	tests := []struct {
		name   string
		action renamer.FileAction
		status string
	}{
		{name: "ready", action: renamer.FileAction{}, status: "● Ready"},
		{name: "duplicate", action: renamer.FileAction{IsDuplicate: true}, status: "⧉ Duplicate"},
		{name: "error", action: renamer.FileAction{IsError: true}, status: "⚠ Error"},
		{name: "skipped", action: renamer.FileAction{IsSkipped: true}, status: "✓ Unchanged"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, _ := actionPresentation(test.action)
			if status != test.status {
				t.Fatalf("status = %q, want %q", status, test.status)
			}
		})
	}
}

func TestProcessLabel(t *testing.T) {
	if got := processLabel(actionCounts{skipped: 3}); got != "Nothing to Change" {
		t.Fatalf("processLabel(skipped) = %q", got)
	}
	if got := processLabel(actionCounts{ready: 1}); got != "Process 1 Photo" {
		t.Fatalf("processLabel(one) = %q", got)
	}
	if got := processLabel(actionCounts{ready: 2, errors: 1}); got != "Process 3 Photos" {
		t.Fatalf("processLabel(many) = %q", got)
	}
}

func TestActionTableHasDraggableHeader(t *testing.T) {
	application := test.NewApp()
	t.Cleanup(application.Quit)

	table, ok := actionTable([]renamer.FileAction{{OriginalPath: "photo.jpg", NewName: "renamed.jpg"}}).(*widget.Table)
	if !ok {
		t.Fatal("actionTable did not return a Fyne table")
	}
	if !table.ShowHeaderRow {
		t.Fatal("actionTable header row is disabled, so columns cannot be dragged")
	}
	if table.CreateHeader == nil || table.UpdateHeader == nil {
		t.Fatal("actionTable does not provide custom column headers")
	}
}
