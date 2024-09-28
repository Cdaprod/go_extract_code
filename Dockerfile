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
RUN go build -o go_extract_code ./main.go

# Stage 2: Create a smaller image
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/go_extract_code .

# Expose port 8080 to the outside world (if applicable)
EXPOSE 8080

# Command to run the executable
CMD ["./go_extract_code"]