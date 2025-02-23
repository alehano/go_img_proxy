# Start with the official Golang image as the build environment
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o link-redirection-service main.go

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/bin/app .

# Copy the configuration file
COPY config.yml .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./app"] 