# Build stage
FROM golang:1.24-alpine AS builder

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

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/ivannovak/glide/pkg/version.Version=${VERSION} \
    -X github.com/ivannovak/glide/pkg/version.BuildDate=${BUILD_DATE} \
    -X github.com/ivannovak/glide/pkg/version.GitCommit=${GIT_COMMIT}" \
    -o glide ./cmd/glide

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates git docker-cli docker-compose

# Create non-root user
RUN addgroup -g 1001 -S glide && \
    adduser -u 1001 -S glide -G glide

# Set working directory
WORKDIR /workspace

# Copy binary from builder
COPY --from=builder /app/glide /usr/local/bin/glide

# Make binary executable
RUN chmod +x /usr/local/bin/glide

# Switch to non-root user
USER glide

# Set entrypoint
ENTRYPOINT ["glide"]
CMD ["--help"]