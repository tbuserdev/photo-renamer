package renamer

import (
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
}

// ScanFiles walks the input folder and generates a list of FileAction for all valid images.
// It does not check for duplicates against the output folder, only generates the new names based on metadata.
func ScanFiles(inputFolder string) ([]FileAction, error) {
	var actions []FileAction

	imageEndings := []string{
		".jpg", ".JPG",
		".jpeg", ".JPEG",
		".png", ".PNG",
		".gif", ".GIF",
		".bmp", ".BMP",
		".tiff", ".TIFF",
		".tif", ".TIF",
		".webp", ".WEBP",
		".heif", ".HEIF",
		".heic", ".HEIC",
		".arw", ".ARW",
		".cr2", ".CR2",
		".cr3", ".CR3",
		".dng", ".DNG",
		".nef", ".NEF",
		".rw2", ".RW2",
		".sr2", ".SR2",
		".srw", ".SRW",
	}

	err := filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if !strings.Contains(path, "@eaDir") && !strings.Contains(path, "Thumbs.db") &&
				!strings.Contains(path, "desktop.ini") && !strings.Contains(path, ".DS_Store") &&
				!strings.Contains(path, "._") && !strings.Contains(path, "._.") &&
				!strings.Contains(path, "Syno") && !strings.Contains(path, "syno") &&
				!strings.Contains(path, "SYNO") && !strings.Contains(path, "Synology") &&
				!strings.Contains(path, "thumb") && !strings.Contains(path, "Thumb") &&
				!strings.Contains(path, "THUMB") && !strings.Contains(path, "Thumbnails") {
				for _, ending := range imageEndings {
					if strings.HasSuffix(path, ending) {
						action := FileAction{
							OriginalPath: path,
						}
						// CREATE NEW FILENAME
						newFileName := Image(path)
						action.NewName = newFileName

						if strings.Contains(newFileName, "error") {
							action.IsError = true
						}

						actions = append(actions, action)
						break
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return actions, nil
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
