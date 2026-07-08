# ─── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download deps first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with optimisations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /app/bin/api ./cmd/api

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/bin/api .

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["./api"]
