package renamer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetExifData returns normalized metadata for the TUI's single-file debug view.
func GetExifData(file string) string {
	items, err := (ExifTool{}).Extract(context.Background(), []string{file})
	if err != nil {
		data, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return string(data)
	}
	if len(items) != 1 {
		return `{"Error":"ExifTool returned no metadata"}`
	}
	item := items[0]
	values := map[string]string{
		"SourceFile": item.SourceFile, "FileType": item.FileType, "FileExtension": item.FileExtension,
		"CaptureTime": item.CaptureTime.Format("2006-01-02T15:04:05Z07:00"), "CaptureSource": item.CaptureSource,
		"Make": item.Make, "Model": item.Model, "Software": item.Software,
	}
	if item.Err != nil {
		values["Error"] = item.Err.Error()
	}
	data, err := json.Marshal(values)
	if err != nil {
		return fmt.Sprintf(`{"Error":%q}`, err.Error())
	}
	return string(data)
}

func filenameFor(metadata Metadata, originalPath string) string {
	if metadata.Err != nil {
		return "METADATA_error"
	}
	ext := filepath.Ext(originalPath)
	if ext == "" {
		return "FILEEXT_error"
	}
	if metadata.CaptureTime.IsZero() {
		return "DATE_error"
	}
	modelName := normalizedModel(metadata.Model)
	makerName := metadata.Make
	editor := editedMetadata(metadata)
	base := metadata.CaptureTime.Format("2006-01-02_15-04-05")
	if makerName != "" && modelName != "" {
		base += "_" + makerName + "-" + modelName
	} else if makerName != "" {
		base += "_" + makerName
	} else if modelName != "" {
		base += "_" + modelName
	}
	if editor != "" {
		base += "_" + editor
	}
	return base + ext
}

// Image is retained for callers that generate a name for one file. Batch preview
// uses ScanFiles and therefore starts only one ExifTool process for the full batch.
func Image(file string) string {
	items, err := (ExifTool{}).Extract(context.Background(), []string{file})
	if err != nil || len(items) != 1 {
		return "METADATA_error"
	}
	return filenameFor(items[0], file)
}

func normalizedModel(value string) string {
	if index := strings.Index(value, "("); index >= 0 {
		return value[:index]
	}
	return value
}

func editedMetadata(metadata Metadata) string {
	software := metadata.Software
	if strings.Contains(software, "Lightroom") || strings.Contains(software, "Adobe Photoshop Lightroom Classic") {
		return "Lightroom"
	}
	if strings.Contains(software, "Photoshop") {
		return "Photoshop"
	}
	if strings.Contains(software, "Photomator") {
		return "Photomator"
	}
	return ""
}

func OpenOutputFolder(folder string) (err error) {
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", "-R", folder).Run()
	case "windows":
		err = exec.Command("explorer", "/select,", folder).Run()
	default:
		log.Printf("unsupported operating system")
	}
	return err
}
