package renamer

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileAction struct {
	OriginalPath  string
	NewName       string
	SourceSize    int64
	SourceModTime time.Time
	SourceHash    [sha256.Size]byte
	IsError       bool
	IsDuplicate   bool
	IsSkipped     bool
	Error         string
}

var ErrPlanStale = errors.New("the folder changed after preview")

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
		if info.IsDir() {
			if isReviewDirectory(inputFolder, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if excludedPath(path) || !supportedExtension(filepath.Ext(path)) {
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
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat scanned file %s: %w", path, err)
		}
		hash, err := fileSHA256(path)
		if err != nil {
			return nil, err
		}
		action := FileAction{
			OriginalPath:  path,
			NewName:       filenameFor(item, path),
			SourceSize:    info.Size(),
			SourceModTime: info.ModTime(),
			SourceHash:    hash,
		}
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

func isReviewDirectory(inputFolder, path string) bool {
	relative, err := filepath.Rel(inputFolder, path)
	if err != nil || filepath.Dir(relative) != "." {
		return false
	}
	switch strings.ToUpper(relative) {
	case "DUPLICATES", "ERROR-OUTPUT":
		return true
	default:
		return false
	}
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
	return PreviewRenameContext(context.Background(), inputFolder, outputFolder)
}

// PreviewRenameContext scans a folder and resolves destination collisions. It
// can be cancelled while ExifTool is running or between filesystem operations.
func PreviewRenameContext(ctx context.Context, inputFolder, outputFolder string) ([]FileAction, error) {
	actions, err := ScanFilesWithExtractor(ctx, inputFolder, ExifTool{})
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
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
				seenPaths[destinationKey(candidateName)] = action.OriginalPath
				break
			}

			collisionPath := seenPaths[destinationKey(candidateName)]
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
				seenPaths[destinationKey(candidateName)] = action.OriginalPath
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
	return RenameContext(context.Background(), actions, outputFolder, errorFolder, duplicateFolder, func(_, _ int) {
		if onProgress != nil {
			onProgress()
		}
	})
}

// RenameContext resolves an unconfirmed batch against the current output folder
// and reports completed and total action counts. Use ApplyPreviewContext when
// actions have already been shown to and confirmed by a user.
func RenameContext(ctx context.Context, actions []FileAction, outputFolder string, errorFolder string, duplicateFolder string, onProgress func(completed, total int)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	resolvedActions, err := resolveCollisions(actions, outputFolder)
	if err != nil {
		return err
	}
	return performRename(ctx, resolvedActions, outputFolder, errorFolder, duplicateFolder, onProgress, true)
}

// ApplyPreviewContext applies exactly the actions shown in a preview. It first
// validates the full plan and aborts with ErrPlanStale before changing any file
// if destinations or source files changed after confirmation.
func ApplyPreviewContext(ctx context.Context, actions []FileAction, outputFolder string, errorFolder string, duplicateFolder string, onProgress func(completed, total int)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePreviewPlan(ctx, actions, outputFolder, errorFolder); err != nil {
		return err
	}
	return performRename(ctx, actions, outputFolder, errorFolder, duplicateFolder, onProgress, false)
}

func performRename(ctx context.Context, actions []FileAction, outputFolder string, errorFolder string, duplicateFolder string, onProgress func(completed, total int), reportSkipped bool) error {
	// Ensure directories exist
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return fmt.Errorf("create output folder: %w", err)
	}
	if err := os.MkdirAll(errorFolder, 0755); err != nil {
		return fmt.Errorf("create error folder: %w", err)
	}
	if err := os.MkdirAll(duplicateFolder, 0755); err != nil {
		return fmt.Errorf("create duplicate folder: %w", err)
	}

	total := len(actions)
	if !reportSkipped {
		total = processableActionCount(actions)
	}

	// Perform renaming
	completed := 0
	for _, action := range actions {
		if err := ctx.Err(); err != nil {
			return err
		}
		if action.IsSkipped {
			if reportSkipped {
				completed++
				reportProgress(onProgress, completed, total)
			}
			continue
		}
		if !reportSkipped {
			if err := validateSource(action); err != nil {
				return err
			}
		}

		if action.IsError {
			originalName := filepath.Base(action.OriginalPath)
			destPath := filepath.Join(errorFolder, originalName)
			if err := moveExact(action.OriginalPath, destPath); err != nil {
				return fmt.Errorf("process %s: %w", originalName, err)
			}
		} else if action.IsDuplicate {
			duplicatePath, err := availablePath(duplicateFolder, "DUPLICATE_"+action.NewName)
			if err != nil {
				return fmt.Errorf("process %s: %w", filepath.Base(action.OriginalPath), err)
			}
			if err := os.Rename(action.OriginalPath, duplicatePath); err != nil {
				return fmt.Errorf("process %s: %w", filepath.Base(action.OriginalPath), err)
			}
		} else {
			destPath := filepath.Join(outputFolder, action.NewName)
			if err := moveExact(action.OriginalPath, destPath); err != nil {
				return fmt.Errorf("process %s: %w", filepath.Base(action.OriginalPath), err)
			}
		}
		completed++
		reportProgress(onProgress, completed, total)
	}

	return nil
}

func validatePreviewPlan(ctx context.Context, actions []FileAction, outputFolder, errorFolder string) error {
	plannedDestinations := make(map[string]string)
	plannedErrorDestinations := make(map[string]struct{})
	for _, action := range actions {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := validateSource(action); err != nil {
			return err
		}
		if action.IsError {
			destination := filepath.Join(errorFolder, filepath.Base(action.OriginalPath))
			key := destinationKey(destination)
			if _, exists := plannedErrorDestinations[key]; exists {
				return fmt.Errorf("%w: error destination %s is reserved more than once", ErrPlanStale, filepath.Base(destination))
			}
			if _, err := os.Stat(destination); err == nil {
				return fmt.Errorf("%w: error destination %s now exists", ErrPlanStale, filepath.Base(destination))
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("validate error destination %s: %w", destination, err)
			}
			plannedErrorDestinations[key] = struct{}{}
			continue
		}

		destination := filepath.Join(outputFolder, action.NewName)
		if action.IsSkipped {
			if !samePath(action.OriginalPath, destination) {
				return fmt.Errorf("%w: %s is no longer unchanged", ErrPlanStale, filepath.Base(action.OriginalPath))
			}
			plannedDestinations[destinationKey(action.NewName)] = action.OriginalPath
			continue
		}
		if action.IsDuplicate {
			collisionPath := plannedDestinations[destinationKey(action.NewName)]
			if collisionPath == "" {
				collisionPath = destination
			}
			equal, err := filesEqual(action.OriginalPath, collisionPath)
			if err != nil || !equal {
				return fmt.Errorf("%w: duplicate status changed for %s", ErrPlanStale, filepath.Base(action.OriginalPath))
			}
			continue
		}
		key := destinationKey(action.NewName)
		if _, exists := plannedDestinations[key]; exists {
			return fmt.Errorf("%w: destination %s is now reserved", ErrPlanStale, action.NewName)
		}
		if _, err := os.Stat(destination); err == nil {
			return fmt.Errorf("%w: destination %s now exists", ErrPlanStale, action.NewName)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("validate destination %s: %w", destination, err)
		}
		plannedDestinations[key] = action.OriginalPath
	}
	return nil
}

// destinationKey is deliberately case-insensitive so a reviewed plan remains
// safe on the default macOS and Windows filesystems.
func destinationKey(path string) string {
	return strings.ToLower(filepath.Clean(path))
}

func validateSource(action FileAction) error {
	info, err := os.Stat(action.OriginalPath)
	if err != nil {
		return fmt.Errorf("%w: source %s is no longer available", ErrPlanStale, filepath.Base(action.OriginalPath))
	}
	if info.Size() != action.SourceSize || !info.ModTime().Equal(action.SourceModTime) {
		return fmt.Errorf("%w: source %s changed", ErrPlanStale, filepath.Base(action.OriginalPath))
	}
	hash, err := fileSHA256(action.OriginalPath)
	if err != nil {
		return fmt.Errorf("%w: source %s cannot be verified", ErrPlanStale, filepath.Base(action.OriginalPath))
	}
	if hash != action.SourceHash {
		return fmt.Errorf("%w: source %s changed", ErrPlanStale, filepath.Base(action.OriginalPath))
	}
	return nil
}

func moveExact(source, destination string) error {
	// Linking is an atomic no-replace operation on the same filesystem. Rename
	// cannot provide this guarantee on Unix because it replaces destination.
	if err := os.Link(source, destination); err != nil {
		if _, statErr := os.Stat(destination); statErr == nil {
			return fmt.Errorf("%w: destination %s now exists", ErrPlanStale, filepath.Base(destination))
		}
		return fmt.Errorf("link %s to %s: %w", source, destination, err)
	}
	if err := os.Remove(source); err != nil {
		rollbackErr := os.Remove(destination)
		if rollbackErr != nil {
			return fmt.Errorf("remove source %s: %w (also failed to remove destination link: %v)", source, err, rollbackErr)
		}
		return fmt.Errorf("remove source %s: %w", source, err)
	}
	if _, err := os.Stat(destination); err != nil {
		return fmt.Errorf("verify destination %s: %w", destination, err)
	}
	return nil
}

func processableActionCount(actions []FileAction) int {
	count := 0
	for _, action := range actions {
		if !action.IsSkipped {
			count++
		}
	}
	return count
}

func reportProgress(callback func(completed, total int), completed, total int) {
	if callback != nil {
		callback(completed, total)
	}
}
