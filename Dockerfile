# Multi-stage build for Render deployment
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Final stage - MongoDB optimized (no postgresql-client)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/bin/server .

# Copy migrations (if needed for MongoDB seeds)
COPY --from=builder /app/migrations ./migrations

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./server"]
