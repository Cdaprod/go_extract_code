// main_test.go
package main

import (
	"testing"
)

func TestGetExtension(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"go", ".go"},
		{"Python", ".py"},
		{"JS", ".js"},
		{"javascript", ".js"},
		{"unknown", ".txt"},
	}

	for _, test := range tests {
		result := getExtension(test.language)
		if result != test.expected {
			t.Errorf("getExtension(%s) = %s; want %s", test.language, result, test.expected)
		}
	}
}

func TestDetermineFilePath(t *testing.T) {
	tests := []struct {
		currentFilePath  string
		currentExtension string
		outputDir        string
		codeLines        []string
		expected         string
	}{
		{
			currentFilePath:  "main.go",
			currentExtension: ".go",
			outputDir:        "output",
			codeLines:        []string{"package main"},
			expected:         "output/main.go",
		},
		{
			currentFilePath:  "",
			currentExtension: ".txt",
			outputDir:        "output",
			codeLines:        []string{"", "Some Content"},
			expected:         "output/default_Some_Content.txt",
		},
	}

	for _, test := range tests {
		result := determineFilePath(test.currentFilePath, test.currentExtension, test.outputDir, test.codeLines)
		if result != test.expected {
			t.Errorf("determineFilePath(%s, %s, %s, %v) = %s; want %s",
				test.currentFilePath, test.currentExtension, test.outputDir, test.codeLines, result, test.expected)
		}
	}
}