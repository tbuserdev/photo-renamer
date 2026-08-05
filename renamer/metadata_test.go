package renamer

import (
	"testing"
	"time"
)

func TestFilenameForMetadata(t *testing.T) {
	captured := time.Date(2023, 4, 19, 19, 17, 54, 0, time.UTC)
	tests := []struct {
		name     string
		metadata Metadata
		expected string
	}{
		{"original", Metadata{CaptureTime: captured, Make: "Canon", Model: "EOS R5"}, "2023-04-19_19-17-54_Canon-EOS R5_Original.jpg"},
		{"camera software", Metadata{CaptureTime: captured, Make: "SONY", Model: "ILCE-7M3", Software: "ILCE-7M3 v2"}, "2023-04-19_19-17-54_SONY-ILCE-7M3.jpg"},
		{"edited", Metadata{CaptureTime: captured, Make: "Google", Model: "Pixel 6 (US)", Software: "Adobe Photoshop"}, "2023-04-19_19-17-54_Google-Pixel 6 _Photoshop.jpg"},
		{"unknown camera", Metadata{CaptureTime: captured}, "2023-04-19_19-17-54_Unknown-Unknown_Original.jpg"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := filenameFor(test.metadata, "photo.jpg"); got != test.expected {
				t.Fatalf("filename = %q; want %q", got, test.expected)
			}
		})
	}
}

func TestFilenameForMetadataErrors(t *testing.T) {
	if got := filenameFor(Metadata{}, "photo.jpg"); got != "DATE_error" {
		t.Fatalf("missing date = %q", got)
	}
	if got := filenameFor(Metadata{CaptureTime: time.Now()}, "photo"); got != "FILEEXT_error" {
		t.Fatalf("missing extension = %q", got)
	}
}
