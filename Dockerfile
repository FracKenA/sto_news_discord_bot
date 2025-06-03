# Multi-architecture Dockerfile for cross-platform builds
# BuildKit optimized approach: Let Docker handle cross-compilation

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# Arguments for target platform (BuildKit provides these automatically)
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

# Arguments for build info
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    git

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary for target architecture
# BuildKit handles cross-compilation automatically when CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT} -linkmode external -extldflags '-static'" \
    -o /app/stobot ./cmd/stobot

# Runtime stage
FROM alpine:latest

# Arguments to determine target platform
ARG TARGETARCH

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 stobot && \
    adduser -D -s /bin/sh -u 1000 -G stobot stobot

# Create data directory
RUN mkdir -p /data && chown stobot:stobot /data

# Copy the binary from builder stage
COPY --from=builder /app/stobot /usr/local/bin/stobot

# Fix permissions
RUN chown stobot:stobot /usr/local/bin/stobot && \
    chmod +x /usr/local/bin/stobot

# Switch to non-root user
USER stobot

# Set working directory
WORKDIR /data

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep stobot || exit 1

# Default command
CMD ["stobot", "--channels-path", "/data/channels.txt", "--database-path", "/data/stobot.db"]
