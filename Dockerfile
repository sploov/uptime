# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o uptime-engine cmd/server/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates only, sqlite-libs not needed)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/uptime-engine .

# Copy example config (user should mount their own)
COPY config.yaml .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./uptime-engine"]
