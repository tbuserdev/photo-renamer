package renamer

import (
	"strings"
	"testing"
)

// Helper function to create a mock JSON string mimicking what gjson expects
func createMockJSON(date, make, model, software string) string {
	// Simple JSON construction
	parts := []string{}
	if date != "" {
		parts = append(parts, `"DateTimeOriginal": "`+date+`"`)
	}
	if make != "" {
		parts = append(parts, `"Make": "`+make+`"`)
	}
	if model != "" {
		parts = append(parts, `"Model": "`+model+`"`)
	}
	if software != "" {
		parts = append(parts, `"Software": "`+software+`"`)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func TestDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{createMockJSON("2023:04:19 19:17:54", "", "", ""), "2023-04-19_19-17-54"},
		{createMockJSON("", "", "", ""), ""},
	}

	for _, test := range tests {
		result := date(test.input)
		if result != test.expected {
			t.Errorf("date(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{createMockJSON("", "", "Canon EOS 5D Mark IV", ""), "Canon EOS 5D Mark IV"},
		{createMockJSON("", "", "ILCE-7M3", ""), "ILCE-7M3"},
		{createMockJSON("", "", "Pixel 6 (US)", ""), "Pixel 6 "}, // Expects truncation before '('
		{createMockJSON("", "", "", ""), "Unknown"},
	}

	for _, test := range tests {
		result := model(test.input)
		if result != test.expected {
			t.Errorf("model(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestMaker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{createMockJSON("", "Canon", "", ""), "Canon"},
		{createMockJSON("", "SONY", "", ""), "SONY"},
		{createMockJSON("", "", "", ""), "Unknown"},
	}

	for _, test := range tests {
		result := maker(test.input)
		if result != test.expected {
			t.Errorf("maker(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestEdited(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Software contains Model -> returns Model
		{createMockJSON("", "", "ILCE-7M3", "ILCE-7M3 v2.0"), "ILCE-7M3"},
		// Software contains "Adobe Lightroom" -> returns "Lightroom"
		{createMockJSON("", "", "Canon", "Adobe Lightroom Classic 10.0"), "Lightroom"},
		// Software contains "Ver.1.0" -> returns Model
		{createMockJSON("", "", "Nikon Z6", "Ver.1.01"), "Nikon Z6"},
		// No match -> returns ""
		{createMockJSON("", "", "Camera", "Unknown Software"), ""},
	}

	for _, test := range tests {
		result := edited(test.input)
		if result != test.expected {
			t.Errorf("edited(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

// Note: TestImage function is harder to unit test directly because it calls openJson which reads a real file.
// We can skip it here and rely on the manual tests or refactor the code later to accept an interface for file reading.
// For now, testing the private helper functions covers the logic complexity.
