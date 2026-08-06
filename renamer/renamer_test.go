package renamer

import (
	"context"
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

	actions, err := ScanFilesWithExtractor(context.Background(), tmpDir, testExtractor{})
	if err != nil {
		t.Fatalf("ScanFiles failed: %v", err)
	}

	// We expect exactly 1 file (test.jpg). The others should be ignored.
	// Note: ScanFiles might try to read metadata and fail, returning an error in the name,
	// but it should still return a FileAction for the valid file.

	if len(actions) != 1 {
		t.Errorf("Expected 1 file action, got %d", len(actions))
		for _, a := range actions {
			t.Logf("Found: %s", a.OriginalPath)
		}
	} else {
		if filepath.Base(actions[0].OriginalPath) != "test.jpg" {
			t.Errorf("Expected test.jpg, got %s", filepath.Base(actions[0].OriginalPath))
		}
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

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
