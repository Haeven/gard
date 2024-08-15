# - TODO: Cargo build rust encoder
# Use the official Golang image as the builder
# Set the working directory for Engine installation

FROM rust:latest AS cargo-builder
WORKDIR /app/encoder

# Copy the Engine module files
COPY ./encoder/ ./

# Set the working directory for Engine installation
RUN cargo install --path .
RUN cargo build --release

# Copy the source code
# COPY ./engine /app/engine

FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /engine

# Use a smaller base image for the final container
FROM ubuntu:22.04

# Copy the compiled binary from the builder stage
COPY --from=builder ./engine /usr/local/bin/engine
COPY --from=cargo-builder /app/encoder /usr/local/bin/encoder

# Create the uploads directory
# RUN mkdir -p /uploads

# Set the working directory
# WORKDIR /uploads

# Set the command to run the application
CMD ["engine"]

# Expose the application port
EXPOSE 8081
