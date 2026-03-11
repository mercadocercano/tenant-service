# ==============================================
# Tenant Service - Optimized Multi-stage Dockerfile
# ==============================================

# ==============================================
# Stage 1: Dependencies and cache optimization
# ==============================================
FROM golang:1.22-alpine AS deps
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependency files and local libs, then download modules
COPY go.mod go.sum ./
COPY libs/eventbus ./libs/eventbus
RUN go mod download && go mod verify

# ==============================================
# Stage 2: Build stage
# ==============================================
FROM deps AS builder

# Copy source code
COPY . .

# Build optimized binary with security hardening
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o tenant-service ./cmd/api

# ==============================================
# Stage 3: Development stage (with Air hot reload)
# ==============================================
FROM golang:1.22-alpine AS development

# Security: Create non-root user first
RUN addgroup -g 1001 -S appgroup && \
    adduser -S -D -h /app -s /bin/sh -G appgroup -u 1001 appuser

# Install runtime dependencies including Air
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    postgresql-client \
    git \
    && cp /usr/share/zoneinfo/UTC /etc/localtime \
    && echo "UTC" > /etc/timezone \
    && apk del tzdata

# Install Air for hot reload (compatible with Go 1.22)
RUN go install github.com/cosmtrek/air@v1.49.0

WORKDIR /app

# Copy go mod files first (for better caching)
COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download

# Copy source code
COPY --chown=appuser:appgroup . .

# Create tmp directory for Air
RUN mkdir -p tmp scripts migrations logs && \
    chown -R appuser:appgroup tmp scripts migrations logs

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8120/health || exit 1

# Expose ports
EXPOSE 8120

# Use Air for hot reload in development
CMD ["air", "-c", ".air.toml"]

# ==============================================
# Stage 4: Production stage (Distroless)
# ==============================================
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Metadata
LABEL org.opencontainers.image.title="Tenant Service" \
      org.opencontainers.image.description="Multi-tenant Configuration Service" \
      org.opencontainers.image.source="https://github.com/mercadocercano/tenant-service" \
      org.opencontainers.image.vendor="SaaS MT Team" \
      org.opencontainers.image.licenses="MIT"

WORKDIR /app

# Copy binary and essential files only
COPY --from=builder --chown=nonroot:nonroot /app/tenant-service ./
COPY --from=builder --chown=nonroot:nonroot /app/scripts ./scripts/
COPY --from=builder --chown=nonroot:nonroot /app/migrations ./migrations/

# Use distroless nonroot user (uid=65532)
USER nonroot

# Expose ports
EXPOSE 8120

ENTRYPOINT ["./tenant-service"]

# ==============================================
# Default stage: Development
# ==============================================
FROM development
