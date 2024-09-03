# Start with the official Golang image to build the Go application
FROM  golang:1.20 AS builder

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# RUN apt-get install libvpx-dev
# Update package list and install dependencies
RUN apt-get update && \
    apt-get install -y \
    build-essential \
    git \
    cmake \
    libvpx-dev \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install vpxenc (VP9 encoder)
RUN cd /tmp && \
    git clone https://github.com/webmproject/libvpx.git && \
    cd libvpx && \
    ./configure && \
    make && \
    make install && \
    ldconfig

# Install Av1an
RUN cd /tmp && \
    git clone https://github.com/master-of-zen/Av1an.git && \
    cd Av1an && \
    make && \
    cp av1an /usr/local/bin/

# Clean up
RUN rm -rf /tmp/*


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for dependency caching
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app with necessary flags for smaller binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server ./cmd

# Start a new stage from scratch (empty base image) for the final image to keep it small
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Copy any additional necessary files (like config.yaml or .env) if needed
COPY config.yaml ./
COPY .env ./

# Expose the port the server will run on
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
