# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')} -X main.commitHash=${COMMIT_HASH:-unknown}" \
    -o xenforo-to-gh-discussions ./cmd/xenforo-to-gh-discussions

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 app && \
    adduser -D -s /bin/sh -u 1000 -G app app

# Copy the binary from builder stage
COPY --from=builder /app/xenforo-to-gh-discussions /usr/local/bin/xenforo-to-gh-discussions

# Set proper permissions
RUN chmod +x /usr/local/bin/xenforo-to-gh-discussions

# Switch to non-root user
USER app

# Set working directory
WORKDIR /home/app

# Expose any ports if needed (uncomment if your app serves HTTP)
# EXPOSE 8080

# Health check (adjust based on your app's capabilities)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD xenforo-to-gh-discussions --help > /dev/null || exit 1

# Default command
ENTRYPOINT ["xenforo-to-gh-discussions"]
CMD ["--help"]
