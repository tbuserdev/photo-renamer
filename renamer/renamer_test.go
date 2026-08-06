package renamer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testExtractor struct{}

func (testExtractor) Extract(_ context.Context, paths []string) ([]Metadata, error) {
	items := make([]Metadata, len(paths))
	for index, path := range paths {
		items[index] = Metadata{SourceFile: path, CaptureTime: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), Make: "Test", Model: "Camera"}
	}
	return items, nil
}

func TestScanFiles_IgnoresExcludedDirectories(t *testing.T) {
	// Create a temp directory structure
	tmpDir, err := os.MkdirTemp("", "photo-renamer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid image file
	validFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(validFile, []byte("fake image data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create excluded directory and file
	excludedDir := filepath.Join(tmpDir, ".DS_Store") // As directory for test case, though usually a file
	if err := os.Mkdir(excludedDir, 0755); err != nil {
		t.Fatal(err)
	}
	excludedFile := filepath.Join(excludedDir, "ignore.jpg")
	if err := os.WriteFile(excludedFile, []byte("fake data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create excluded directory "Thumbnails"
	thumbDir := filepath.Join(tmpDir, "Thumbnails")
	if err := os.Mkdir(thumbDir, 0755); err != nil {
		t.Fatal(err)
	}
	thumbFile := filepath.Join(thumbDir, "thumb.jpg")
	if err := os.WriteFile(thumbFile, []byte("fake thumb data"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DUPLICATES", "ERROR-OUTPUT"} {
		reviewDir := filepath.Join(tmpDir, name)
		if err := os.Mkdir(reviewDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(reviewDir, "review.jpg"), []byte("review data"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	nestedReviewDir := filepath.Join(tmpDir, "album", "DUPLICATES")
	if err := os.MkdirAll(nestedReviewDir, 0755); err != nil {
		t.Fatal(err)
	}
	nestedFile := filepath.Join(nestedReviewDir, "nested.jpg")
	writeTestFile(t, nestedFile, "nested photo")

	actions, err := ScanFilesWithExtractor(context.Background(), tmpDir, testExtractor{})
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// Direct review-output folders are ignored, while a nested user folder that
	// happens to share the same name remains part of the scan.
	// Note: ScanFiles might try to read metadata and fail, returning an error in the name,
	// but it should still return a FileAction for the valid file.

	if len(actions) != 2 {
		t.Errorf("Expected 2 file actions, got %d", len(actions))
		for _, a := range actions {
			t.Logf("Found: %s", a.OriginalPath)
		}
	}
	found := map[string]bool{}
	for _, action := range actions {
		found[action.OriginalPath] = true
	}
	if !found[validFile] || !found[nestedFile] {
		t.Errorf("scanned paths = %v, want %s and %s", found, validFile, nestedFile)
	}
}

func TestScanFiles_ValidExtensions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "photo-renamer-ext-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := []string{"test.jpg", "test.PNG", "test.arw", "clip.MOV", "clip.mp4", "test.txt", "test.pdf"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	actions, err := ScanFilesWithExtractor(context.Background(), tmpDir, testExtractor{})
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 5 // jpg, PNG, arw, MOV, mp4
	if len(actions) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(actions))
	}
}

func TestResolveCollisions_DifferentContentGetsSuffix(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	existing := filepath.Join(outputDir, "2024-01-02_03-04-05_Test-Camera.jpg")
	incoming := filepath.Join(inputDir, "incoming.jpg")
	writeTestFile(t, existing, "existing photo")
	writeTestFile(t, incoming, "different photo")

	actions, err := resolveCollisions([]FileAction{{
		OriginalPath: incoming,
		NewName:      filepath.Base(existing),
	}}, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	if actions[0].IsDuplicate {
		t.Fatal("different content was classified as a duplicate")
	}
	if got, want := actions[0].NewName, "2024-01-02_03-04-05_Test-Camera_2.jpg"; got != want {
		t.Fatalf("NewName = %q, want %q", got, want)
	}
}

func TestResolveCollisions_IdenticalContentIsDuplicate(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	existing := filepath.Join(outputDir, "photo.jpg")
	incoming := filepath.Join(inputDir, "incoming.jpg")
	writeTestFile(t, existing, "same photo bytes")
	writeTestFile(t, incoming, "same photo bytes")

	actions, err := resolveCollisions([]FileAction{{OriginalPath: incoming, NewName: "photo.jpg"}}, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	if !actions[0].IsDuplicate {
		t.Fatal("identical content was not classified as a duplicate")
	}
	if got, want := actions[0].NewName, "photo.jpg"; got != want {
		t.Fatalf("NewName = %q, want %q", got, want)
	}
}

func TestResolveCollisions_SameBatchUsesHashesAndSuffixes(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	first := filepath.Join(inputDir, "first.jpg")
	second := filepath.Join(inputDir, "second.jpg")
	third := filepath.Join(inputDir, "third.jpg")
	writeTestFile(t, first, "first photo")
	writeTestFile(t, second, "second photo")
	writeTestFile(t, third, "first photo")

	actions, err := resolveCollisions([]FileAction{
		{OriginalPath: first, NewName: "photo.jpg"},
		{OriginalPath: second, NewName: "photo.jpg"},
		{OriginalPath: third, NewName: "photo.jpg"},
	}, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	if actions[0].NewName != "photo.jpg" || actions[0].IsDuplicate {
		t.Fatalf("unexpected first action: %+v", actions[0])
	}
	if actions[1].NewName != "photo_2.jpg" || actions[1].IsDuplicate {
		t.Fatalf("unexpected second action: %+v", actions[1])
	}
	if actions[2].NewName != "photo.jpg" || !actions[2].IsDuplicate {
		t.Fatalf("unexpected third action: %+v", actions[2])
	}
}

func TestRename_PreservesDifferentContentCollision(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	duplicateDir := filepath.Join(outputDir, "DUPLICATES")
	errorDir := filepath.Join(outputDir, "ERROR-OUTPUT")
	existing := filepath.Join(outputDir, "photo.jpg")
	incoming := filepath.Join(inputDir, "incoming.jpg")
	writeTestFile(t, existing, "existing photo")
	writeTestFile(t, incoming, "different photo")

	err := Rename([]FileAction{{OriginalPath: incoming, NewName: "photo.jpg"}}, outputDir, errorDir, duplicateDir, func() {})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(outputDir, "photo_2.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "different photo"; got != want {
		t.Fatalf("renamed content = %q, want %q", got, want)
	}
	if entries, err := os.ReadDir(duplicateDir); err != nil || len(entries) != 0 {
		t.Fatalf("duplicate directory entries = %v, err = %v", entries, err)
	}
}

func TestRenameContext_CancelledBeforeStartDoesNotCreateFolders(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "output")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RenameContext(ctx, nil, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if err != context.Canceled {
		t.Fatalf("RenameContext error = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatalf("cancelled rename created output folder: %v", err)
	}
}

func TestRenameContext_ReportsCompletedAndTotal(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	first := filepath.Join(inputDir, "first.jpg")
	second := filepath.Join(inputDir, "second.jpg")
	writeTestFile(t, first, "first")
	writeTestFile(t, second, "second")

	var progress [][2]int
	err := RenameContext(context.Background(), []FileAction{
		{OriginalPath: first, NewName: "first-renamed.jpg"},
		{OriginalPath: second, NewName: "second-renamed.jpg"},
	}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), func(completed, total int) {
		progress = append(progress, [2]int{completed, total})
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]int{{1, 2}, {2, 2}}
	if len(progress) != len(want) || progress[0] != want[0] || progress[1] != want[1] {
		t.Fatalf("progress = %v, want %v", progress, want)
	}
}

func TestApplyPreviewContext_RejectsStaleDestinationBeforeChanges(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	incoming := filepath.Join(inputDir, "incoming.jpg")
	destination := filepath.Join(outputDir, "renamed.jpg")
	writeTestFile(t, incoming, "incoming")
	writeTestFile(t, destination, "appeared after preview")

	action := previewAction(t, incoming, "renamed.jpg")
	err := ApplyPreviewContext(context.Background(), []FileAction{action}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
	if content, err := os.ReadFile(incoming); err != nil || string(content) != "incoming" {
		t.Fatalf("source changed after stale plan: content=%q err=%v", content, err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "ERROR-OUTPUT")); !os.IsNotExist(err) {
		t.Fatalf("stale plan created review folders: %v", err)
	}
}

func TestApplyPreviewContext_ProgressExcludesSkippedActions(t *testing.T) {
	outputDir := t.TempDir()
	unchanged := filepath.Join(outputDir, "unchanged.jpg")
	incomingDir := t.TempDir()
	incoming := filepath.Join(incomingDir, "incoming.jpg")
	writeTestFile(t, unchanged, "unchanged")
	writeTestFile(t, incoming, "incoming")

	var progress [][2]int
	skipped := previewAction(t, unchanged, "unchanged.jpg")
	skipped.IsSkipped = true
	action := previewAction(t, incoming, "renamed.jpg")
	err := ApplyPreviewContext(context.Background(), []FileAction{skipped, action}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), func(completed, total int) {
		progress = append(progress, [2]int{completed, total})
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]int{{1, 1}}
	if len(progress) != 1 || progress[0] != want[0] {
		t.Fatalf("progress = %v, want %v", progress, want)
	}
}

func TestApplyPreviewContext_RejectsChangedSourceBeforeChanges(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	source := filepath.Join(inputDir, "incoming.jpg")
	writeTestFile(t, source, "original")
	action := previewAction(t, source, "renamed.jpg")
	writeTestFile(t, source, "changed source with a different size")

	err := ApplyPreviewContext(context.Background(), []FileAction{action}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "renamed.jpg")); !os.IsNotExist(err) {
		t.Fatalf("stale source was moved: %v", err)
	}
}

func TestApplyPreviewContext_RejectsSameSizeChangedSource(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	source := filepath.Join(inputDir, "incoming.jpg")
	writeTestFile(t, source, "original")
	action := previewAction(t, source, "renamed.jpg")
	writeTestFile(t, source, "modified")
	if err := os.Chtimes(source, action.SourceModTime, action.SourceModTime); err != nil {
		t.Fatal(err)
	}

	err := ApplyPreviewContext(context.Background(), []FileAction{action}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
}

func TestApplyPreviewContext_RejectsOccupiedErrorDestination(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	errorDir := filepath.Join(outputDir, "ERROR-OUTPUT")
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(inputDir, "bad.jpg")
	writeTestFile(t, source, "bad input")
	writeTestFile(t, filepath.Join(errorDir, "bad.jpg"), "existing error")
	action := previewAction(t, source, "error_bad.jpg")
	action.IsError = true

	err := ApplyPreviewContext(context.Background(), []FileAction{action}, outputDir, errorDir, filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
	if content, err := os.ReadFile(source); err != nil || string(content) != "bad input" {
		t.Fatalf("source changed after stale error plan: content=%q err=%v", content, err)
	}
}

func TestApplyPreviewContext_RejectsRepeatedErrorDestination(t *testing.T) {
	root := t.TempDir()
	outputDir := t.TempDir()
	first := filepath.Join(root, "album-a", "bad.jpg")
	second := filepath.Join(root, "album-b", "bad.jpg")
	if err := os.MkdirAll(filepath.Dir(first), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(second), 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, first, "first")
	writeTestFile(t, second, "second")
	firstAction := previewAction(t, first, "error_bad.jpg")
	firstAction.IsError = true
	secondAction := previewAction(t, second, "error_bad.jpg")
	secondAction.IsError = true

	err := ApplyPreviewContext(context.Background(), []FileAction{firstAction, secondAction}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
	if _, err := os.Stat(first); err != nil {
		t.Fatalf("first source changed after rejected plan: %v", err)
	}
	if _, err := os.Stat(second); err != nil {
		t.Fatalf("second source changed after rejected plan: %v", err)
	}
}

func TestApplyPreviewContext_RejectsCaseVariantErrorDestination(t *testing.T) {
	root := t.TempDir()
	outputDir := t.TempDir()
	first := filepath.Join(root, "album-a", "bad.jpg")
	second := filepath.Join(root, "album-b", "BAD.JPG")
	if err := os.MkdirAll(filepath.Dir(first), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(second), 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, first, "first")
	writeTestFile(t, second, "second")
	firstAction := previewAction(t, first, "error_bad.jpg")
	firstAction.IsError = true
	secondAction := previewAction(t, second, "error_BAD.JPG")
	secondAction.IsError = true

	err := ApplyPreviewContext(context.Background(), []FileAction{firstAction, secondAction}, outputDir, filepath.Join(outputDir, "ERROR-OUTPUT"), filepath.Join(outputDir, "DUPLICATES"), nil)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("ApplyPreviewContext error = %v, want ErrPlanStale", err)
	}
}

func TestMoveExact_DoesNotReplaceDestination(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.jpg")
	destination := filepath.Join(dir, "destination.jpg")
	writeTestFile(t, source, "source")
	writeTestFile(t, destination, "destination")

	err := moveExact(source, destination)
	if !errors.Is(err, ErrPlanStale) {
		t.Fatalf("moveExact error = %v, want ErrPlanStale", err)
	}
	if content, err := os.ReadFile(destination); err != nil || string(content) != "destination" {
		t.Fatalf("destination replaced: content=%q err=%v", content, err)
	}
	if content, err := os.ReadFile(source); err != nil || string(content) != "source" {
		t.Fatalf("source changed: content=%q err=%v", content, err)
	}
}

func previewAction(t *testing.T, path, newName string) FileAction {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return FileAction{
		OriginalPath:  path,
		NewName:       newName,
		SourceSize:    info.Size(),
		SourceModTime: info.ModTime(),
		SourceHash:    mustFileHash(t, path),
	}
}

func mustFileHash(t *testing.T, path string) [32]byte {
	t.Helper()
	hash, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
