# Krustron - Multi-stage Containerfile
# Author: Anubhav Gain <anubhavg@infopercept.com>
# Build: podman build -t krustron:latest .

# ============================================================================
# Stage 1: Build Go backend
# ============================================================================
FROM docker.io/golang:1.22-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') \
    -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown') \
    -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/krustron ./cmd/krustron

# ============================================================================
# Stage 2: Build React frontend
# ============================================================================
FROM docker.io/node:20-alpine AS frontend-builder

WORKDIR /app/web/dashboard

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy package files
COPY web/dashboard/package.json web/dashboard/pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY web/dashboard/ ./

# Build frontend
RUN pnpm run build

# ============================================================================
# Stage 3: Final runtime image
# ============================================================================
FROM docker.io/alpine:3.19

LABEL org.opencontainers.image.title="Krustron"
LABEL org.opencontainers.image.description="Kubernetes-Native Platform for CI/CD, GitOps, and Cluster Management"
LABEL org.opencontainers.image.vendor="Anubhav Gain"
LABEL org.opencontainers.image.source="https://github.com/anubhavg-icpl/krustron"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 krustron && \
    adduser -u 1000 -G krustron -s /bin/sh -D krustron

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/krustron /app/krustron

# Copy frontend assets
COPY --from=frontend-builder /app/web/dashboard/dist /app/web/static

# Copy config template
COPY config.yaml /app/config.yaml.template

# Set ownership
RUN chown -R krustron:krustron /app

# Switch to non-root user
USER krustron

# Expose ports
EXPOSE 8080 8443 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["/app/krustron"]
CMD ["serve"]
