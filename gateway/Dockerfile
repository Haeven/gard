# Start with the official Golang image to build the Go application
FROM golang:1.20-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app/gateway

# Copy go.mod and go.sum files first for dependency caching
COPY ./gateway/go.mod ./gateway/go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the rest of the application code
COPY ./gateway .

# Build the Go app with necessary flags for smaller binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Start a new stage from scratch (empty base image) for the final image to keep it small
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/gateway/server .

# Copy any additional necessary files (like config.yaml or .env) if needed
COPY config.yaml ./
COPY .env ./

# Expose the port the server will run on
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
