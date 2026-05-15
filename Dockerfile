# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /registration-tool .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

# Install CA certs (needed for TLS/SMTP) and create non-root user
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app \
    && adduser  -S app -G app

WORKDIR /app

# Copy binary and templates
COPY --from=builder /registration-tool .
COPY templates/ templates/

# Drop privileges
USER app

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["./registration-tool"]
