# Multi-stage build for smaller image
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o let-monitor-go ./cmd/app

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/let-monitor-go .
COPY --from=builder /app/config.example.json .

# Create data directory
RUN mkdir -p /app/data

EXPOSE 5556

CMD ["./let-monitor-go"]