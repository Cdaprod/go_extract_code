// cli.go
package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"flag"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

// Colorized print functions
var (
	purple = color.New(color.FgMagenta).Add(color.Bold)
	blue   = color.New(color.FgBlue)
	green  = color.New(color.FgGreen)
	orange = color.New(color.FgHiYellow)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed).Add(color.Bold)
)

// printHeader prints a purple-colored header message.
func printHeader(message string) {
	purple.Println("\n=== " + message + " ===\n")
}

// printStep prints a blue-colored step message.
func printStep(message string) {
	blue.Println("➡️  " + message)
}

// printSuccess prints a green-colored success message.
func printSuccess(message string) {
	green.Println("✅  " + message)
}

// printWarning prints a yellow-colored warning message.
func printWarning(message string) {
	yellow.Println("⚠️  " + message)
}

// printError prints a red-colored error message.
func printError(message string) {
	red.Println("❌  " + message)
}

// promptConfirmation asks the user to confirm an action.
func promptConfirmation(prompt string) bool {
	orange.Print(prompt + " [y/n]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// runCLI orchestrates the CLI flow.
func runCLI() {
	printHeader("Enhanced Code Block Extractor CLI")

	// Define command-line flags
	buildFilePath := flag.String("build", "BUILD.md", "Path to the BUILD.md file")
	outputDir := flag.String("out", ".", "Output working directory where files will be created")
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
		time.Sleep(100 * time.Millisecond)
	}

	bar.Finish()
	printSuccess(fmt.Sprintf("Processing complete! %d files created.", len(fileDataList)))
}

// Entry point for the program
func main() {
	runCLI()
}