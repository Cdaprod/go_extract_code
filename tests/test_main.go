// main_test.go
package main

import (
	"os"
	"strings"
	"testing"
)

// TestProcess tests the Process method of FileProcessor, including the handling of nested code blocks.
func TestProcess(t *testing.T) {
	// Create a temporary BUILD.md file with various code blocks, including nested ones
	buildContent := `
# BUILD.md

## Sample Code Block

**file:** \`go_web_server_app/main.go\`
\`\`\`go
package main

func main() {
    // Entry point
}
\`\`\`

**file:** \`go_web_server_app/routes.go\`
\`\`\`go
package main

func initializeRoutes() {
    // Initialize routes
}
\`\`\`

# Nested Code Block Example (This should be ignored)
\`\`\`markdown
# Nested Markdown File Example
\`\`\`python
# This is a nested code block within a markdown code block
def ignored_function():
    print("This should not be processed as a file.")
\`\`\`
\`\`\`
This nested markdown code block should be ignored entirely.
\`\`\`
`

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "BUILD.md")
	if err != nil {
		t.Fatalf("Failed to create temp BUILD.md: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the content to the temporary file
	_, err = tmpFile.WriteString(buildContent)
	if err != nil {
		t.Fatalf("Failed to write to temp BUILD.md: %v", err)
	}
	tmpFile.Close()

	// Initialize FileProcessor with the temporary BUILD.md file
	fp := NewFileProcessor(tmpFile.Name(), "output_dir", false)
	fileDataList, err := fp.Process()
	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// Define the expected files and their contents
	expectedFiles := map[string][]string{
		"go_web_server_app/main.go": {
			"package main",
			"",
			"func main() {",
			"    // Entry point",
			"}",
		},
		"go_web_server_app/routes.go": {
			"package main",
			"",
			"func initializeRoutes() {",
			"    // Initialize routes",
			"}",
		},
	}

	// Verify that the correct number of files were processed
	if len(fileDataList) != len(expectedFiles) {
		t.Errorf("Process() returned %d files; want %d", len(fileDataList), len(expectedFiles))
	}

	// Check that each expected file was processed correctly
	for _, fileData := range fileDataList {
		expectedContent, exists := expectedFiles[fileData.Path]
		if !exists {
			t.Errorf("Unexpected file path: %s", fileData.Path)
			continue
		}

		// Compare the content of the files line by line
		if len(fileData.Contents) != len(expectedContent) {
			t.Errorf("File %s has %d lines; want %d", fileData.Path, len(fileData.Contents), len(expectedContent))
			continue
		}

		for i, line := range expectedContent {
			if fileData.Contents[i] != line {
				t.Errorf("File %s, Line %d: got '%s'; want '%s'", fileData.Path, i+1, fileData.Contents[i], line)
			}
		}
	}

	// Check that nested code blocks were not processed
	for _, fileData := range fileDataList {
		if strings.Contains(fileData.Path, "nested") {
			t.Errorf("Nested code block was incorrectly processed: %s", fileData.Path)
		}
	}
}