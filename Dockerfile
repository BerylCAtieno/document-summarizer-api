# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o /app/server ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy any static files (if needed)
COPY --from=builder /app/.env ./

# Expose port
EXPOSE 8080

# Run the app
CMD ["./server"]
