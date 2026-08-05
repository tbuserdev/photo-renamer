package renamer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileAction struct {
	OriginalPath string
	NewName      string
	IsError      bool
	IsDuplicate  bool
	IsSkipped    bool
	Error        string
}

// ScanFiles walks the input folder and generates a list of FileAction for all valid images.
// It does not check for duplicates against the output folder, only generates the new names based on metadata.
func ScanFiles(inputFolder string) ([]FileAction, error) {
	return ScanFilesWithExtractor(context.Background(), inputFolder, ExifTool{})
}

func ScanFilesWithExtractor(ctx context.Context, inputFolder string, extractor MetadataExtractor) ([]FileAction, error) {
	var paths []string
	err := filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || excludedPath(path) || !supportedExtension(filepath.Ext(path)) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	metadata, err := extractor.Extract(ctx, paths)
	if err != nil {
		return nil, err
	}
	if len(metadata) != len(paths) {
		return nil, fmt.Errorf("metadata extractor returned %d records for %d files", len(metadata), len(paths))
	}
	actions := make([]FileAction, 0, len(paths))
	for index, path := range paths {
		item := metadata[index]
		action := FileAction{OriginalPath: path, NewName: filenameFor(item, path)}
		if item.Err != nil {
			action.IsError = true
			action.Error = item.Err.Error()
		} else if strings.Contains(action.NewName, "error") {
			action.IsError = true
			action.Error = action.NewName
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func supportedExtension(extension string) bool {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".webp",
		".heif", ".heic", ".arw", ".cr2", ".cr3", ".dng", ".nef", ".rw2", ".sr2", ".srw",
		".mov", ".mp4":
		return true
	default:
		return false
	}
}

func excludedPath(path string) bool {
	lower := strings.ToLower(path)
	for _, fragment := range []string{"@eadir", "thumbs.db", "desktop.ini", ".ds_store", "._", "syno", "thumb"} {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

func PreviewRename(inputFolder, outputFolder string) ([]FileAction, error) {
	actions, err := ScanFiles(inputFolder)
	if err != nil {
		return nil, err
	}

	seenNames := make(map[string]bool)
	for i, action := range actions {
		if !action.IsError {
			// Check if filename is same as proposed
			if filepath.Base(action.OriginalPath) == action.NewName {
				actions[i].IsSkipped = true
				seenNames[action.NewName] = true
				continue
			}

			// CHECK FOR DUPLICATES
			// 1. Check if file exists in output folder
			_, err := os.Stat(filepath.Join(outputFolder, action.NewName))
			existsOnDisk := !os.IsNotExist(err)

			// 2. Check if we already saw this name in this batch
			seenInBatch := seenNames[action.NewName]

			if existsOnDisk || seenInBatch {
				actions[i].IsDuplicate = true
			} else {
				seenNames[action.NewName] = true
			}
		}
	}
	return actions, nil
}

func moveFile(src, dest, duplicateDir string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return os.Rename(src, dest)
	}
	// Move to duplicate folder
	filename := filepath.Base(dest)
	return os.Rename(src, filepath.Join(duplicateDir, "DUPLICATE_"+filename))
}

func Rename(actions []FileAction, outputFolder string, errorFolder string, duplicateFolder string, onProgress func()) error {
	// Ensure directories exist
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(errorFolder, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(duplicateFolder, 0755); err != nil {
		return err
	}

	// Perform renaming
	for _, action := range actions {
		if action.IsSkipped {
			onProgress()
			continue
		}

		if action.IsError {
			originalName := filepath.Base(action.OriginalPath)
			destPath := filepath.Join(errorFolder, originalName)
			if err := moveFile(action.OriginalPath, destPath, duplicateFolder); err != nil {
				return err
			}
		} else {
			destPath := filepath.Join(outputFolder, action.NewName)
			if err := moveFile(action.OriginalPath, destPath, duplicateFolder); err != nil {
				return err
			}
		}
		onProgress()
	}

	return nil
}
