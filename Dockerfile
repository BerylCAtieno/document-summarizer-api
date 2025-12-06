FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go binary
RUN go build -o server ./cmd/server

FROM alpine:latest

# Required for sqlite (if needed)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations folder from builder
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

# Optional: copy any other static assets your app needs
# COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Run the app
CMD ["./server"]
