package renamer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
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
	return resolveCollisions(actions, outputFolder)
}

func resolveCollisions(actions []FileAction, outputFolder string) ([]FileAction, error) {
	seenPaths := make(map[string]string)
	for i, action := range actions {
		if action.IsError {
			continue
		}

		actions[i].IsDuplicate = false
		actions[i].IsSkipped = false
		baseName := action.NewName
		for suffix := 1; ; suffix++ {
			candidateName := suffixedName(baseName, suffix)
			candidatePath := filepath.Join(outputFolder, candidateName)

			if samePath(action.OriginalPath, candidatePath) {
				actions[i].NewName = candidateName
				actions[i].IsSkipped = true
				seenPaths[candidateName] = action.OriginalPath
				break
			}

			collisionPath := seenPaths[candidateName]
			if collisionPath == "" {
				info, err := os.Stat(candidatePath)
				switch {
				case err == nil && !info.IsDir():
					collisionPath = candidatePath
				case err == nil:
					return nil, fmt.Errorf("rename destination is a directory: %s", candidatePath)
				case !os.IsNotExist(err):
					return nil, fmt.Errorf("check rename destination %s: %w", candidatePath, err)
				}
			}

			if collisionPath == "" {
				actions[i].NewName = candidateName
				seenPaths[candidateName] = action.OriginalPath
				break
			}

			equal, err := filesEqual(action.OriginalPath, collisionPath)
			if err != nil {
				return nil, err
			}
			if equal {
				actions[i].NewName = candidateName
				actions[i].IsDuplicate = true
				break
			}
		}
	}
	return actions, nil
}

func suffixedName(name string, suffix int) string {
	if suffix <= 1 {
		return name
	}
	extension := filepath.Ext(name)
	return strings.TrimSuffix(name, extension) + fmt.Sprintf("_%d", suffix) + extension
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func filesEqual(left, right string) (bool, error) {
	leftInfo, err := os.Stat(left)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", left, err)
	}
	rightInfo, err := os.Stat(right)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", right, err)
	}
	if leftInfo.Size() != rightInfo.Size() {
		return false, nil
	}

	leftHash, err := fileSHA256(left)
	if err != nil {
		return false, err
	}
	rightHash, err := fileSHA256(right)
	if err != nil {
		return false, err
	}
	return leftHash == rightHash, nil
}

func fileSHA256(path string) ([sha256.Size]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("open %s for duplicate check: %w", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("hash %s: %w", path, err)
	}
	var sum [sha256.Size]byte
	copy(sum[:], hash.Sum(nil))
	return sum, nil
}

func moveFile(src, dest, duplicateDir string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return os.Rename(src, dest)
	}
	// Move to duplicate folder
	filename := filepath.Base(dest)
	duplicatePath, err := availablePath(duplicateDir, "DUPLICATE_"+filename)
	if err != nil {
		return err
	}
	return os.Rename(src, duplicatePath)
}

func availablePath(folder, name string) (string, error) {
	for suffix := 1; ; suffix++ {
		candidate := filepath.Join(folder, suffixedName(name, suffix))
		_, err := os.Stat(candidate)
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("check destination %s: %w", candidate, err)
		}
	}
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

	resolvedActions, err := resolveCollisions(actions, outputFolder)
	if err != nil {
		return err
	}

	// Perform renaming
	for _, action := range resolvedActions {
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
		} else if action.IsDuplicate {
			duplicatePath, err := availablePath(duplicateFolder, "DUPLICATE_"+action.NewName)
			if err != nil {
				return err
			}
			if err := os.Rename(action.OriginalPath, duplicatePath); err != nil {
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
