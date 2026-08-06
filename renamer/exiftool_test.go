package renamer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseExifToolJSONPreservesGroupedTagsAndOffset(t *testing.T) {
	input := `[{"SourceFile":"phone.heic","File:FileType":"HEIC","File:FileTypeExtension":"HEIC","EXIF:DateTimeOriginal":"2024:03:10 01:30:00-05:00","QuickTime:CreateDate":"2024:03:10 06:30:00Z","EXIF:Make":"Apple","EXIF:Model":"iPhone 15","XMP:Software":"Photoshop"}]`

	items, err := parseExifToolJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	got := items[0]
	if got.CaptureSource != "EXIF:DateTimeOriginal" {
		t.Fatalf("capture source = %q", got.CaptureSource)
	}
	if got.CaptureTime.Format(time.RFC3339) != "2024-03-10T01:30:00-05:00" {
		t.Fatalf("capture time = %s", got.CaptureTime.Format(time.RFC3339))
	}
	if got.FileType != "HEIC" || got.Make != "Apple" || got.Model != "iPhone 15" || got.Software != "Photoshop" {
		t.Fatalf("unexpected normalized metadata: %#v", got)
	}
}

func TestParseExifToolJSONRecognizesRealExifToolGroupNames(t *testing.T) {
	input := `[{"SourceFile":"camera.arw","File:FileType":"ARW","ExifIFD:DateTimeOriginal":"2026:07:18 07:29:27","IFD0:Make":"SONY","IFD0:Model":"ILCE-7M3","IFD0:Software":"ILCE-7M3 v4.01"}]`

	items, err := parseExifToolJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	got := items[0]
	if got.Err != nil || got.CaptureSource != "ExifIFD:DateTimeOriginal" || got.Make != "SONY" || got.Model != "ILCE-7M3" {
		t.Fatalf("unexpected normalized metadata: %#v", got)
	}
}

func TestCaptureTimeFallsBackToFilenameThenFileModificationDate(t *testing.T) {
	input := `[{"SourceFile":"DUPLICATE_2025-08-03_16-22-56_Unknown.jpg","System:FileModifyDate":"2026:01:02 03:04:05+01:00"},{"SourceFile":"IMG_8735.jpg","System:FileModifyDate":"2025:04:13 20:09:36+02:00"}]`

	items, err := parseExifToolJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if items[0].CaptureSource != "FileName" || items[0].CaptureTime.Format("2006-01-02 15:04:05") != "2025-08-03 16:22:56" {
		t.Fatalf("filename fallback = %#v", items[0])
	}
	if items[1].CaptureSource != "System:FileModifyDate" || items[1].CaptureTime.Format(time.RFC3339) != "2025-04-13T20:09:36+02:00" {
		t.Fatalf("file modification fallback = %#v", items[1])
	}
}

func TestVideoTimestampPrecedenceAndQuickTimeUTC(t *testing.T) {
	input := `[{"SourceFile":"clip.mov","File:FileType":"MOV","QuickTime:CreationDate":"2024:06:01 12:00:00+02:00","QuickTime:CreateDate":"2024:06:01 10:00:00","QuickTime:MediaCreateDate":"2024:06:01 09:00:00"},{"SourceFile":"clip.mp4","File:FileType":"MP4","QuickTime:CreationDate":"2024:06:01 08:00:00","QuickTime:CreateDate":"2024:06:01 10:00:00"}]`

	items, err := parseExifToolJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if items[0].CaptureSource != "QuickTime:CreationDate" || items[0].CaptureTime.Format(time.RFC3339) != "2024-06-01T12:00:00+02:00" {
		t.Fatalf("explicit offset not preferred: %#v", items[0])
	}
	if items[1].CaptureTime.Location() != time.UTC || items[1].CaptureTime.Format(time.RFC3339) != "2024-06-01T10:00:00Z" {
		t.Fatalf("timezone-less QuickTime date not interpreted as UTC: %#v", items[1])
	}
}

func TestRepresentativePhoneAndCameraMetadata(t *testing.T) {
	input := `[{"SourceFile":"iphone.heic","File:FileType":"HEIC","EXIF:DateTimeOriginal":"2024:01:02 03:04:05-08:00","EXIF:Make":"Apple","EXIF:Model":"iPhone 15"},{"SourceFile":"iphone.mov","File:FileType":"MOV","Keys:CreationDate":"2024:01:02 03:04:05-08:00","QuickTime:Make":"Apple","QuickTime:Model":"iPhone 15"},{"SourceFile":"android.heif","File:FileType":"HEIF","EXIF:DateTimeOriginal":"2024:02:03 04:05:06+01:00","EXIF:Make":"Google","EXIF:Model":"Pixel 8"},{"SourceFile":"android.mp4","File:FileType":"MP4","QuickTime:CreateDate":"2024:02:03 03:05:06","QuickTime:Make":"Google","QuickTime:Model":"Pixel 8"},{"SourceFile":"camera.mov","File:FileType":"MOV","QuickTime:MediaCreateDate":"2024:03:04 05:06:07","QuickTime:Make":"SONY","QuickTime:Model":"ILCE-7M4"}]`
	items, err := parseExifToolJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	for index, item := range items {
		if item.Err != nil || item.CaptureTime.IsZero() {
			t.Fatalf("fixture %d did not normalize: %#v", index, item)
		}
		if got := filenameFor(item, item.SourceFile); strings.Contains(got, "error") {
			t.Fatalf("fixture %d did not produce filename: %s", index, got)
		}
	}
}

func TestParseExifToolJSONReportsMalformedPartialAndMissingMetadata(t *testing.T) {
	if _, err := parseExifToolJSON([]byte(`[{`)); err == nil || !strings.Contains(err.Error(), "malformed ExifTool JSON") {
		t.Fatalf("malformed JSON error = %v", err)
	}
	items, err := parseExifToolJSON([]byte(`[{"SourceFile":"bad.heic","Error":"File format error"},{"SourceFile":"empty.mov","File:FileType":"MOV"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(items[0].Err.Error(), "File format error") {
		t.Fatalf("corrupt file error = %v", items[0].Err)
	}
	if !errors.Is(items[1].Err, ErrCaptureTimeMissing) {
		t.Fatalf("missing capture error = %v", items[1].Err)
	}
}

func TestExifToolBatchUsesOneProcessAndSafeArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := filepath.Join(dir, "fake-exiftool")
	scriptBody := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$ARGS_FILE\"\nprintf '[{\"SourceFile\":\"%s\",\"File:FileType\":\"HEIC\",\"EXIF:DateTimeOriginal\":\"2024:01:02 03:04:05\"},{\"SourceFile\":\"%s\",\"File:FileType\":\"MOV\",\"QuickTime:CreateDate\":\"2024:01:02 03:04:05\"}]' \"$SOURCE_ONE\" \"$SOURCE_TWO\"\n"
	if err := os.WriteFile(script, []byte(scriptBody), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARGS_FILE", argsFile)
	paths := []string{filepath.Join(dir, "space ü.heic"), filepath.Join(dir, "-leading.mov")}
	t.Setenv("SOURCE_ONE", paths[0])
	t.Setenv("SOURCE_TWO", paths[1])

	items, err := (ExifTool{Path: script}).Extract(context.Background(), paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSpace(string(args)), "\n")
	if got[len(got)-3] != "--" || got[len(got)-2] != paths[0] || got[len(got)-1] != paths[1] {
		t.Fatalf("unsafe/unexpected args: %#v", got)
	}
}

func TestExifToolIgnoresSuccessfulProcessDiagnosticsOnStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	script := filepath.Join(t.TempDir(), "fake-exiftool")
	body := "#!/bin/sh\nprintf '[{\"SourceFile\":\"photo.jpg\",\"File:FileType\":\"JPEG\",\"EXIF:DateTimeOriginal\":\"2024:01:02 03:04:05\"}]'\necho '1 image files read' >&2\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}

	items, err := (ExifTool{Path: script}).Extract(context.Background(), []string{"photo.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Err != nil {
		t.Fatalf("unexpected batch result: %#v", items)
	}
}

func TestExifToolActionableErrorsAndCancellation(t *testing.T) {
	_, err := (ExifTool{Path: filepath.Join(t.TempDir(), "missing")}).Extract(context.Background(), []string{"x.heic"})
	if err == nil || !strings.Contains(err.Error(), "install ExifTool") {
		t.Fatalf("missing dependency error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = (ExifTool{Path: "exiftool"}).Extract(ctx, []string{"x.heic"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestExifToolReportsProcessFailureAndPartialOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	dir := t.TempDir()
	failing := filepath.Join(dir, "failing-exiftool")
	if err := os.WriteFile(failing, []byte("#!/bin/sh\necho 'decoder exploded' >&2\nexit 9\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := (ExifTool{Path: failing}).Extract(context.Background(), []string{"x.mov"}); err == nil || !strings.Contains(err.Error(), "decoder exploded") {
		t.Fatalf("process failure = %v", err)
	}

	partial := filepath.Join(dir, "partial-exiftool")
	if err := os.WriteFile(partial, []byte("#!/bin/sh\nprintf '[{\"SourceFile\":\"one.mov\",\"File:FileType\":\"MOV\",\"QuickTime:CreateDate\":\"2024:01:02 03:04:05\"}]'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	_, err := (ExifTool{Path: partial}).Extract(context.Background(), []string{"one.mov", "two.mov"})
	if err == nil || !strings.Contains(err.Error(), "partial JSON") {
		t.Fatalf("partial output error = %v", err)
	}
}

func TestExifToolReturnsCorruptInputAsPerFileError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	script := filepath.Join(t.TempDir(), "corrupt-exiftool")
	body := "#!/bin/sh\nprintf '[{\"SourceFile\":\"good.mov\",\"File:FileType\":\"MOV\",\"QuickTime:CreateDate\":\"2024:01:02 03:04:05\"},{\"SourceFile\":\"bad.mov\",\"ExifTool:Error\":\"File format error\"}]'\nexit 1\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	items, err := (ExifTool{Path: script}).Extract(context.Background(), []string{"good.mov", "bad.mov"})
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Err != nil || items[1].Err == nil || !strings.Contains(items[1].Err.Error(), "File format error") {
		t.Fatalf("unexpected batch result: %#v", items)
	}
}

func TestFilenameUsesCaptureWallTimeWithoutDiscardingOffset(t *testing.T) {
	captured, err := parseMetadataTime("2024:03:10 01:30:00-05:00", false)
	if err != nil {
		t.Fatal(err)
	}
	metadata := Metadata{CaptureTime: captured, Make: "Apple", Model: "iPhone 15"}
	if got := filenameFor(metadata, "photo.heic"); got != "2024-03-10_01-30-00_Apple-iPhone 15.heic" {
		t.Fatalf("filename = %q", got)
	}
	if metadata.CaptureTime.Format(time.RFC3339) != "2024-03-10T01:30:00-05:00" {
		t.Fatalf("offset was discarded: %s", metadata.CaptureTime.Format(time.RFC3339))
	}
}
