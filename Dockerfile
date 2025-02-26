# Use the official Golang image for building the application
FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
WORKDIR /app/cmd/main
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

# Use a minimal Debian-based image for the final container to support TUI
FROM debian:bookworm-slim

WORKDIR /root/
RUN apt-get update && apt-get install -y xterm
COPY --from=builder /app/cmd/main/server .
EXPOSE 13000 13100
CMD ["/server"]
