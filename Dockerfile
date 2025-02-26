# Use the official Golang image for building the application
FROM golang:1.24 as builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the entire source code
COPY . ./

# Ensure we are in the correct directory before building
#RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./cmd/main/main.go
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/main/main.go

# Use a minimal Debian-based image for the final container to support TUI
FROM debian:bookworm-slim

WORKDIR /root/
# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /

# Expose necessary ports
EXPOSE 13000 13100

# Keep the container running for debugging
CMD ["/server"]

