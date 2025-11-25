# Multi-stage build for smaller image
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o let-monitor-go ./cmd/app

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/let-monitor-go .

# Copy template files
COPY internal/server/templates ./internal/server/templates

# Create data directory
RUN mkdir -p /app/data

EXPOSE 5556

CMD ["./let-monitor-go"]