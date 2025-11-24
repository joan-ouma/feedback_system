.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build the application
build:
	go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Database migrations
migrate-up:
	psql ${DATABASE_URL} < migrations/001_initial_schema.up.sql

migrate-down:
	psql ${DATABASE_URL} < migrations/001_initial_schema.down.sql

# Docker commands (if using Docker)
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

