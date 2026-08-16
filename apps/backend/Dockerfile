# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build tools & git
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY . .

# Build API and migration binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/migrate ./cmd/migrate

# Production Run Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Create non-root unprivileged app user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binaries & config
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/acis-config.yaml /app/acis-config.yaml

# Copy migration SQL scripts from ../migrations to /app/migrations
COPY ../migrations /app/migrations

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

CMD ["/app/api"]
