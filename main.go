// main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileProcessor handles the processing of the build file.
type FileProcessor struct {
	BuildFilePath string
	OutputDir     string
	Verbose       bool
}

// NewFileProcessor initializes a new FileProcessor.
func NewFileProcessor(buildFilePath, outputDir string, verbose bool) *FileProcessor {
	return &FileProcessor{
		BuildFilePath: buildFilePath,
		OutputDir:     outputDir,
		Verbose:       verbose,
	}
}

// Process reads the build file and extracts code blocks.
func (fp *FileProcessor) Process() ([]FileData, error) {
	file, err := os.Open(fp.BuildFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", fp.BuildFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regular expressions
	filePathRegex := regexp.MustCompile(`^\*\*file:\*\*\s*` + "`" + `([^` + "`" + `]+)` + "`")
	codeBlockRegex := regexp.MustCompile("^```([a-zA-Z]+)")

	var currentFilePath string
	var currentExtension string
	inCodeBlock := false
	var codeLines []string
	var fileDataList []FileData

	for scanner.Scan() {
		line := scanner.Text()

		if !inCodeBlock {
			// Check for file path
			matches := filePathRegex.FindStringSubmatch(line)
			if len(matches) == 2 {
				currentFilePath = matches[1]
				if fp.Verbose {
					fmt.Printf("[DEBUG] Detected file path: %s\n", currentFilePath)
				}
				continue
			}

			// Check for code block start
			codeMatches := codeBlockRegex.FindStringSubmatch(line)
			if len(codeMatches) == 2 {
				currentExtension = getExtension(codeMatches[1])
				inCodeBlock = true
				codeLines = []string{}
				if fp.Verbose {
					fmt.Printf("[DEBUG] Starting %s code block\n", codeMatches[1])
				}
				continue
			}
		} else {
			// Check for code block end
			if strings.TrimSpace(line) == "```" {
				// Determine file path and extension
				fullPath := determineFilePath(currentFilePath, currentExtension, fp.OutputDir, codeLines)

				fileData := FileData{
					Path:     fullPath,
					Contents: codeLines,
				}
				fileDataList = append(fileDataList, fileData)

				// Reset state
				inCodeBlock = false
				currentFilePath = ""
				continue
			} else {
				// Accumulate code lines
				codeLines = append(codeLines, line)

				// If first line inside code block and no file path, check for comment
				if len(codeLines) == 1 && currentFilePath == "" {
					possiblePath := extractPathFromComment(line, currentExtension)
					if possiblePath != "" {
						currentFilePath = possiblePath
						if fp.Verbose {
							fmt.Printf("[DEBUG] Detected path from comment: %s\n", currentFilePath)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", fp.BuildFilePath, err)
	}

	return fileDataList, nil
}

// FileData holds information about a file to be created.
type FileData struct {
	Path     string
	Contents []string
}

// getExtension maps language names to file extensions.
func getExtension(language string) string {
	extensions := map[string]string{
		"go":        ".go",
		"python":    ".py",
		"js":        ".js",
		"javascript": ".js",
		"html":      ".html",
		"markdown":  ".md",
		"yaml":      ".yaml",
		"json":      ".json",
		"shell":     ".sh",
		"bash":      ".sh",
		// Add more languages and extensions as needed
	}
	if ext, exists := extensions[strings.ToLower(language)]; exists {
		return ext
	}
	return ".txt" // Default extension
}

// determineFilePath constructs the full file path.
func determineFilePath(currentFilePath, currentExtension, outputDir string, codeLines []string) string {
	if currentFilePath != "" {
		// Ensure the path has the correct extension
		ext := filepath.Ext(currentFilePath)
		if ext == "" {
			currentFilePath += currentExtension
		}
		return filepath.Join(outputDir, currentFilePath)
	}

	// Default to outputDir's parent with a generated filename
	defaultFileName := fmt.Sprintf("default_%s%s", strings.ReplaceAll(firstNonEmptyLine(codeLines), " ", "_"), currentExtension)
	return filepath.Join(outputDir, "..", defaultFileName)
}

// firstNonEmptyLine returns the first non-empty line from codeLines.
func firstNonEmptyLine(codeLines []string) string {
	for _, line := range codeLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return "unknown"
}

// extractPathFromComment extracts file path from a comment line.
func extractPathFromComment(line, extension string) string {
	// Define comment styles for different languages
	commentPatterns := map[string]*regexp.Regexp{
		"//":     regexp.MustCompile(`^//\s*(.+)$`),        // Go, JavaScript
		"#":      regexp.MustCompile(`^#\s*(.+)$`),         // Python, Shell
		"<!--":   regexp.MustCompile(`^<!--\s*(.+)\s*-->$`), // HTML comments
		"/*":     regexp.MustCompile(`^/\*\s*(.+)\s*\*/$`), // C-style comments
		"--":     regexp.MustCompile(`^--\s*(.+)$`),        // SQL
	}

	for prefix, regex := range commentPatterns {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			matches := regex.FindStringSubmatch(line)
			if len(matches) == 2 {
				// Ensure the path has the correct extension
				ext := filepath.Ext(matches[1])
				if ext == "" {
					return matches[1] + extension
				}
				return matches[1]
			}
		}
	}

	return ""
}

// validateOutputDir ensures the output directory exists or creates it.
func validateOutputDir(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// Attempt to create the directory
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("error accessing output directory %s: %w", dir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

// writeCodeToFile creates and writes the code to the specified file.
// If the file already exists, it appends a numeric suffix to the filename (e.g., file(2).ext).
func writeCodeToFile(path string, codeLines []string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", path, err)
		}
	}

	// Generate a unique file path if the file already exists
	finalPath := generateUniqueFilePath(path)

	// Create or overwrite the file
	file, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", finalPath, err)
	}
	defer file.Close()

	// Write the code lines to the file
	writer := bufio.NewWriter(file)
	for _, line := range codeLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", finalPath, err)
		}
	}

	// Flush the buffer to ensure all data is written
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush buffer for file %s: %w", finalPath, err)
	}

	return nil
}

// generateUniqueFilePath checks if a file exists and appends a numeric suffix if necessary (e.g., file(2).ext).
func generateUniqueFilePath(path string) string {
	// Check if the file already exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// File exists, generate a unique path by appending a number
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	counter := 2

	for {
		// Generate the new path with a counter
		newPath := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}