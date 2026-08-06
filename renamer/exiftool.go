package renamer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var ErrCaptureTimeMissing = errors.New("capture timestamp is missing")

// Metadata is the normalized boundary between ExifTool and filename generation.
// CaptureSource retains the metadata group so duplicate short tag names remain distinct.
type Metadata struct {
	SourceFile    string
	CaptureTime   time.Time
	CaptureSource string
	Make          string
	Model         string
	Software      string
	FileType      string
	FileExtension string
	Err           error
}

type MetadataExtractor interface {
	Extract(context.Context, []string) ([]Metadata, error)
}

// ExifTool extracts one batch with one process. Path defaults to exiftool on PATH.
type ExifTool struct {
	Path string
}

func (e ExifTool) Extract(ctx context.Context, paths []string) ([]Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}
	path := e.Path
	if path == "" {
		path = "exiftool"
	}
	args := []string{
		"-j", "-G1", "-a", "-api", "QuickTimeUTC=1",
		"-FileType", "-FileTypeExtension", "-DateTimeOriginal", "-CreateDate",
		"-CreationDate", "-MediaCreateDate", "-TrackCreateDate", "-FileModifyDate", "-Make", "-Model", "-Software", "--",
	}
	args = append(args, paths...)
	command := exec.CommandContext(ctx, path, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	items, outputErr := validateExifToolOutput(output, len(paths))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("ExifTool is required but was not found; install ExifTool and ensure `exiftool` is on PATH: %w", err)
		}
		if outputErr == nil && hasPerFileError(items) {
			return items, nil
		}
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(string(output))
			if detail == "" {
				detail = err.Error()
			}
		}
		return nil, fmt.Errorf("ExifTool failed: %s", detail)
	}
	if outputErr != nil {
		return nil, outputErr
	}
	return items, nil
}

func validateExifToolOutput(output []byte, expected int) ([]Metadata, error) {
	items, err := parseExifToolJSON(output)
	if err != nil {
		return nil, fmt.Errorf("ExifTool produced unusable output; verify the ExifTool installation and version: %w", err)
	}
	if len(items) != expected {
		return nil, fmt.Errorf("ExifTool returned partial JSON; verify the ExifTool installation and input files: got %d records for %d files", len(items), expected)
	}
	return items, nil
}

func hasPerFileError(items []Metadata) bool {
	for _, item := range items {
		if item.Err != nil {
			return true
		}
	}
	return false
}

func parseExifToolJSON(data []byte) ([]Metadata, error) {
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("malformed ExifTool JSON: %w", err)
	}
	items := make([]Metadata, 0, len(records))
	for _, record := range records {
		item := Metadata{
			SourceFile:    stringTag(record, "SourceFile"),
			FileType:      firstTag(record, "File:FileType", "FileType"),
			FileExtension: firstTag(record, "File:FileTypeExtension", "FileTypeExtension"),
			Make:          groupedTag(record, "Make", "IFD0", "ExifIFD", "EXIF", "QuickTime", "XMP", "XMP-tiff"),
			Model:         groupedTag(record, "Model", "IFD0", "ExifIFD", "EXIF", "QuickTime", "XMP", "XMP-tiff"),
			Software:      groupedTag(record, "Software", "IFD0", "ExifIFD", "EXIF", "XMP", "XMP-xmp", "QuickTime"),
		}
		if message := groupedTag(record, "Error", "ExifTool", "File"); message != "" {
			item.Err = fmt.Errorf("ExifTool could not read %s: %s", filepath.Base(item.SourceFile), message)
			items = append(items, item)
			continue
		}
		item.CaptureTime, item.CaptureSource, item.Err = captureTime(record, isVideo(item.FileType, item.FileExtension, item.SourceFile))
		if item.Err != nil {
			item.CaptureTime, item.CaptureSource, item.Err = fallbackCaptureTime(record, item.SourceFile)
		}
		items = append(items, item)
	}
	return items, nil
}

func captureTime(record map[string]any, video bool) (time.Time, string, error) {
	type candidate struct {
		key              string
		quickTime        bool
		requiresTimezone bool
	}
	var candidates []candidate
	if video {
		candidates = []candidate{
			{"QuickTime:CreationDate", true, true}, {"Keys:CreationDate", false, true},
			{"QuickTime:CreateDate", true, false}, {"QuickTime:MediaCreateDate", true, false},
			{"QuickTime:TrackCreateDate", true, false},
		}
	} else {
		candidates = []candidate{
			{"ExifIFD:DateTimeOriginal", false, false}, {"EXIF:DateTimeOriginal", false, false},
			{"XMP-exif:DateTimeOriginal", false, false}, {"XMP:DateTimeOriginal", false, false},
			{"QuickTime:CreationDate", true, false}, {"Keys:CreationDate", false, false},
			{"ExifIFD:CreateDate", false, false}, {"EXIF:CreateDate", false, false},
			{"XMP-xmp:CreateDate", false, false}, {"XMP:CreateDate", false, false},
			{"QuickTime:CreateDate", true, false},
		}
	}
	for _, candidate := range candidates {
		value := stringTag(record, candidate.key)
		if value == "" || (candidate.requiresTimezone && !hasExplicitTimezone(value)) {
			continue
		}
		parsed, err := parseMetadataTime(value, candidate.quickTime)
		if err != nil {
			continue
		}
		return parsed, candidate.key, nil
	}
	return time.Time{}, "", ErrCaptureTimeMissing
}

var filenameTimestamp = regexp.MustCompile(`(?:^|_)(\d{4}-\d{2}-\d{2})_(\d{2})(?:-?)(\d{2})(?:-?)(\d{2})(?:_|$)`)

func fallbackCaptureTime(record map[string]any, sourceFile string) (time.Time, string, error) {
	if match := filenameTimestamp.FindStringSubmatch(filepath.Base(sourceFile)); match != nil {
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", match[1]+" "+match[2]+":"+match[3]+":"+match[4], time.Local)
		if err == nil {
			return parsed, "FileName", nil
		}
	}
	if value := stringTag(record, "System:FileModifyDate"); value != "" {
		if parsed, err := parseMetadataTime(value, false); err == nil {
			return parsed, "System:FileModifyDate", nil
		}
	}
	return time.Time{}, "", ErrCaptureTimeMissing
}

func hasExplicitTimezone(value string) bool {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "Z") {
		return true
	}
	if len(value) < 6 {
		return false
	}
	suffix := value[len(value)-6:]
	return (suffix[0] == '+' || suffix[0] == '-') && suffix[3] == ':'
}

func parseMetadataTime(value string, quickTime bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		"2006:01:02 15:04:05.999999999Z07:00", "2006:01:02 15:04:05Z07:00",
		"2006:01:02 15:04:05.999999999-07:00", "2006:01:02 15:04:05-07:00",
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	location := time.Local
	if quickTime {
		// QuickTimeUTC=1 asks ExifTool to express integer QuickTime dates as UTC.
		location = time.UTC
	}
	for _, layout := range []string{"2006:01:02 15:04:05.999999999", "2006:01:02 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported metadata timestamp %q", value)
}

func firstTag(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringTag(record, key); value != "" {
			return value
		}
	}
	return ""
}

func groupedTag(record map[string]any, name string, groups ...string) string {
	keys := make([]string, 0, len(groups)+1)
	for _, group := range groups {
		keys = append(keys, group+":"+name)
	}
	return firstTag(record, append(keys, name)...)
}

func stringTag(record map[string]any, key string) string {
	value, ok := record[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}

func isVideo(fileType, extension, path string) bool {
	value := strings.ToLower(firstNonEmpty(fileType, extension, strings.TrimPrefix(filepath.Ext(path), ".")))
	switch value {
	case "mov", "mp4", "m4v":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
