# Stage 1: Build the Go binary
FROM golang:1.20 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o go_extract_code ./cli.go ./main.go

# Stage 2: Create a smaller image and copy the binary from the builder stage
FROM alpine:latest

# Set the working directory in the new container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/go_extract_code .

# Expose port if your application listens to a specific port
# EXPOSE 8080

# Run the binary
CMD ["./go_extract_code"]