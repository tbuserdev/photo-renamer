package renamer

import (
	"os"
	"path/filepath"
	"testing"
)

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

	actions, err := ScanFiles(tmpDir)
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

	files := []string{"test.jpg", "test.PNG", "test.arw", "test.txt", "test.pdf"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	actions, err := ScanFiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 3 // jpg, PNG, arw
	if len(actions) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(actions))
	}
}
