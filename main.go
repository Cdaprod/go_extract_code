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
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

// FileProcessor handles the processing of the build file.
type FileProcessor struct {
	BuildFilePath string
	OutputDir     string
	Verbose       bool
}

// FileData holds information about a file to be created.
type FileData struct {
	Path     string
	Contents []string
}

// NewFileProcessor initializes a new FileProcessor.
func NewFileProcessor(buildFilePath, outputDir string, verbose bool) *FileProcessor {
	return &FileProcessor{
		BuildFilePath: buildFilePath,
		OutputDir:     outputDir,
		Verbose:       verbose,
	}
}

// Process reads the build file and extracts top-level code blocks.
func (fp *FileProcessor) Process() ([]FileData, error) {
	file, err := os.Open(fp.BuildFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", fp.BuildFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regular expressions
	filePathRegex := regexp.MustCompile(`^\*\*file:\*\*\s*` + "`" + `([^` + "`" + `]+)` + "`")
	codeBlockStartRegex := regexp.MustCompile("^```([a-zA-Z]+)")
	codeBlockEndRegex := regexp.MustCompile("^```$")

	var currentFilePath string
	var currentExtension string
	inCodeBlock := false
	codeLines := []string{}
	fileDataList := []FileData{}

	for scanner.Scan() {
		line := scanner.Text()

		if codeBlockStartRegex.MatchString(line) {
			if inCodeBlock {
				// Nested code block detected; skip processing
				fp.verbosePrint("[DEBUG] Nested code block detected. Skipping...")
				continue
			}
			// Starting a new top-level code block
			inCodeBlock = true
			matches := codeBlockStartRegex.FindStringSubmatch(line)
			if len(matches) == 2 {
				currentExtension = getExtension(matches[1])
			}
			fp.verbosePrint(fmt.Sprintf("[DEBUG] Starting %s code block", currentExtension))
			continue
		}

		if codeBlockEndRegex.MatchString(line) {
			if inCodeBlock {
				// Ending the top-level code block
				fullPath := determineFilePath(currentFilePath, currentExtension, fp.OutputDir, codeLines)
				fileData := FileData{
					Path:     fullPath,
					Contents: codeLines,
				}
				fileDataList = append(fileDataList, fileData)

				// Reset state
				inCodeBlock = false
				currentFilePath = ""
				codeLines = []string{}
				fp.verbosePrint("[DEBUG] Ending code block")
			}
			continue
		}

		if inCodeBlock {
			// Accumulate code lines only if inside a top-level code block
			codeLines = append(codeLines, line)
		} else {
			// Check for file path specification
			matches := filePathRegex.FindStringSubmatch(line)
			if len(matches) == 2 {
				currentFilePath = matches[1]
				fp.verbosePrint(fmt.Sprintf("[DEBUG] Detected file path: %s", currentFilePath))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", fp.BuildFilePath, err)
	}

	return fileDataList, nil
}

// getExtension maps language names to file extensions.
func getExtension(language string) string {
	extensions := map[string]string{
		"go":          ".go",
		"python":      ".py",
		"js":          ".js",
		"javascript":  ".js",
		"html":        ".html",
		"markdown":    ".md",
		"yml":		   ".yml",
		"yaml":        ".yaml",
		"json":        ".json",
		"shell":       ".sh",
		"bash":        ".sh",
		"csharp":      ".cs",
		"sql":         ".sql",
		"typescript":  ".ts",
		"jsx":         ".jsx",
		"tsx":         ".tsx",
		"graphql":     ".graphql",
		"dockerfile":  ".dockerfile",
		"makefile":    ".mk",
		"powershell":  ".ps1",
		"ruby":        ".rb",
		"perl":        ".pl",
		"lua":         ".lua",
		"scala":       ".scala",
		"elixir":      ".ex",
		"erlang":      ".erl",
		"haskell":     ".hs",
		"clojure":     ".clj",
		"fsharp":      ".fs",
		"r":           ".r",
		"matlab":      ".m",
		"groovy":      ".groovy",
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

	// Default to outputDir with a generated filename
	defaultFileName := fmt.Sprintf("default_%s%s", firstNonEmptyLine(codeLines), currentExtension)
	return filepath.Join(outputDir, defaultFileName)
}

// firstNonEmptyLine returns the first non-empty, sanitized line from codeLines.
func firstNonEmptyLine(codeLines []string) string {
	for _, line := range codeLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Replace spaces with underscores and remove non-alphanumeric characters
			processed := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(trimmed, "_")
			return processed
		}
	}
	return "unknown"
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

// verbosePrint prints messages if verbose mode is enabled.
func (fp *FileProcessor) verbosePrint(message string) {
	if fp.Verbose {
		fmt.Println(message)
	}
}

// runCLI orchestrates the CLI flow.
func runCLI() {
	// Colorized print functions
	var (
		purple = color.New(color.FgMagenta).Add(color.Bold)
		blue   = color.New(color.FgBlue)
		green  = color.New(color.FgGreen)
		orange = color.New(color.FgHiYellow)
		yellow = color.New(color.FgYellow)
		red    = color.New(color.FgRed).Add(color.Bold)
	)

	// Helper functions for colored output
	printHeader := func(message string) {
		purple.Println("\n=== " + message + " ===\n")
	}

	printStep := func(message string) {
		blue.Println("➡️  " + message)
	}

	printSuccess := func(message string) {
		green.Println("✅  " + message)
	}

	printWarning := func(message string) {
		yellow.Println("⚠️  " + message)
	}

	printError := func(message string) {
		red.Println("❌  " + message)
	}

	promptConfirmation := func(prompt string) bool {
		orange.Print(prompt + " [y/n]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		return response == "y" || response == "yes"
	}

	printHeader("Enhanced Code Block Extractor CLI")

	// Define command-line flags
	buildFilePath := flag.String("build", "docs/BUILD.md", "Path to the BUILD.md file")
	outputDir := flag.String("out", "build_output/go_web_server_app", "Output directory where files will be created")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	// Display selected options
	printStep(fmt.Sprintf("Selected Build File: %s", *buildFilePath))
	printStep(fmt.Sprintf("Selected Output Directory: %s", *outputDir))

	// Confirm with the user
	if !promptConfirmation("Do you wish to continue with these settings?") {
		printError("Operation aborted by user.")
		os.Exit(1)
	}

	// Validate output directory
	printStep("Validating output directory...")
	err := validateOutputDir(*outputDir)
	if err != nil {
		printError(fmt.Sprintf("Invalid output directory: %v", err))
		os.Exit(1)
	}
	printSuccess("Output directory is valid.")

	// Process the build file
	printStep("Processing build file...")
	fileProcessor := NewFileProcessor(*buildFilePath, *outputDir, *verbose)
	fileDataList, err := fileProcessor.Process()
	if err != nil {
		printError(fmt.Sprintf("Error processing build file: %v", err))
		os.Exit(1)
	}

	if len(fileDataList) == 0 {
		printWarning("No code blocks found to process.")
		os.Exit(0)
	}

	// Initialize progress bar
	bar := pb.Full.Start(len(fileDataList))
	bar.SetTemplate(`{{string . "prefix"}}{{ bar . "[" "#" "-" "]" }} {{percent .}} | {{counters .}} | {{etime .}}`)
	bar.Set("prefix", "📄 Processing Files: ")

	// Process each file with progress bar
	for _, fileData := range fileDataList {
		// Write the file
		err := writeCodeToFile(fileData.Path, fileData.Contents)
		if err != nil {
			printWarning(fmt.Sprintf("Failed to write file %s: %v", fileData.Path, err))
		} else {
			green.Printf("✅ Created: %s\n", fileData.Path)
		}
		bar.Increment()
		// Simulate processing time
		time.Sleep(10 * time.Millisecond) // Reduced sleep time for faster processing
	}

	bar.Finish()
	printSuccess(fmt.Sprintf("Processing complete! %d files created.", len(fileDataList)))
}

func main() {
	runCLI()
}