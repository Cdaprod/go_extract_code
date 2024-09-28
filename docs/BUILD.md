# BUILD.md

# Go Web Server Application Build Guide

This document provides instructions and code snippets required to build a Go-based web server application. The application includes the main entry point, route definitions, a Dockerfile for containerization, a GitHub Actions workflow for building the binary, and a README for project documentation.

## Project Structure

The project will have the following structure:

```
go_web_server_app/
├── main.go
├── routes.go
├── Dockerfile
├── .github/
│   └── workflows/
│       └── build-binary.yml
└── README.md
```

## File Descriptions and Code Snippets

### 1. `main.go`

**Path:** `go_web_server_app/main.go`

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	// Initialize routes
	routes := initializeRoutes()

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", routes); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
```

### 2. `routes.go`

**Path:** `go_web_server_app/routes.go`

```go
package main

import (
	"fmt"
	"net/http"
)

func initializeRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/hello", helloHandler)
	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Go Web Server!")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!\n", name)
}
```

### 3. `Dockerfile`

**Path:** `go_web_server_app/Dockerfile`

```dockerfile
# Stage 1: Build the Go binary
FROM golang:1.20 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o go_web_server_app ./main.go ./routes.go

# Stage 2: Create a smaller image
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/go_web_server_app .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./go_web_server_app"]
```

### 4. GitHub Actions Workflow: `build-binary.yml`

**Path:** `go_web_server_app/.github/workflows/build-binary.yml`

```yaml
name: Build Go Binary

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      # Step 1: Checkout the repository
      - name: Checkout Repository
        uses: actions/checkout@v3

      # Step 2: Set up Go environment
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'

      # Step 3: Install Dependencies and Update go.sum
      - name: Install Go Dependencies
        run: go mod tidy

      # Step 4: Clean up old binary in the root directory
      - name: Clean up old binary
        run: |
          rm -f go_web_server_app

      # Step 5: Build the Go binary and place it in the root
      - name: Build Go Binary
        run: |
          go build -o go_web_server_app ./main.go ./routes.go

      # Step 6: Commit and push the binary back to the repository
      - name: Commit and Push Binary
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

          # Check if the binary exists
          if [ -f go_web_server_app ]; then
            # Add the binary and go.mod/go.sum to the Git index
            git add go_web_server_app go.mod go.sum

            # Check if there are changes to commit
            if ! git diff --cached --quiet; then
              git commit -m "ci: Add latest built binary go_web_server_app and update go.mod/go.sum [skip ci]"
              git push origin HEAD:main  # Change to your target branch if different
            else
              echo "No changes to commit."
            fi
          else
            echo "Binary not found, skipping commit."
          fi
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 5. `README.md`

**Path:** `go_web_server_app/README.md`

```markdown
# Go Web Server Application

## Overview

This project is a simple Go-based web server application designed to demonstrate how to set up a Go project with routing, Docker containerization, and automated builds using GitHub Actions.

## Features

- **Simple Routing**: Handles basic HTTP routes.
- **Dockerized**: Easily containerize the application using Docker.
- **Continuous Integration**: Automatically build and push binaries using GitHub Actions.

## Project Structure

```
go_web_server_app/
├── main.go
├── routes.go
├── Dockerfile
├── .github/
│   └── workflows/
│       └── build-binary.yml
└── README.md
```

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.20 or later installed
- [Docker](https://www.docker.com/get-started) installed
- [GitHub CLI](https://cli.github.com/) (optional, for GitHub Actions setup)

### Building the Application

1. **Clone the Repository**

   ```bash
   git clone https://github.com/yourusername/go_web_server_app.git
   cd go_web_server_app
   ```

2. **Install Dependencies**

   ```bash
   go mod tidy
   ```

3. **Build the Binary**

   ```bash
   go build -o go_web_server_app ./main.go ./routes.go
   ```

4. **Run the Application**

   ```bash
   ./go_web_server_app
   ```

   The server will start on `http://localhost:8080`.

### Dockerizing the Application

1. **Build the Docker Image**

   ```bash
   docker build -t go_web_server_app .
   ```

2. **Run the Docker Container**

   ```bash
   docker run --rm -p 8080:8080 go_web_server_app
   ```

   The server will be accessible at `http://localhost:8080`.
```