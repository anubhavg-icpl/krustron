# Krustron Dockerfile
# Author: Anubhav Gain <anubhavg@infopercept.com>
#
# Multi-stage build for minimal image size

# Build stage
FROM golang:1.22-alpine AS builder

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

# Build arguments
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s \
        -X github.com/anubhavg-icpl/krustron/pkg/version.Version=${VERSION} \
        -X github.com/anubhavg-icpl/krustron/pkg/version.GitCommit=${GIT_COMMIT} \
        -X github.com/anubhavg-icpl/krustron/pkg/version.BuildTime=${BUILD_TIME}" \
    -o /krustron ./cmd/krustron

# Final stage
FROM alpine:3.19

# Labels
LABEL org.opencontainers.image.title="Krustron"
LABEL org.opencontainers.image.description="Kubernetes-native platform for CI/CD, GitOps, and Cluster Management"
LABEL org.opencontainers.image.authors="Anubhav Gain <anubhavg@infopercept.com>"
LABEL org.opencontainers.image.source="https://github.com/anubhavg-icpl/krustron"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 krustron && \
    adduser -u 1000 -G krustron -s /bin/sh -D krustron

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /krustron /app/krustron

# Copy config
COPY config.yaml /app/config.yaml

# Set ownership
RUN chown -R krustron:krustron /app

# Switch to non-root user
USER krustron

# Expose ports
EXPOSE 8080 50051

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["/app/krustron"]
CMD ["serve"]
