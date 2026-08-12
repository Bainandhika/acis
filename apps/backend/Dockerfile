# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Run Stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-ca-certificates ca-certificates tzdata

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/acis-config.yaml /app/acis-config.yaml

EXPOSE 8080

CMD ["/app/api"]
