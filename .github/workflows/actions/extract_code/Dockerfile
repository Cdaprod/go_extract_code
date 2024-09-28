# Use an official Go runtime as a parent image
FROM golang:1.20

# Set the working directory inside the container
WORKDIR /app

# Copy the action code into the container
COPY . .

# Build the Go binary
RUN go mod tidy && go build -o /go_extract_code ./main.go ./cli.go

# Run the binary with the inputs provided by the action
ENTRYPOINT ["/go_extract_code"]