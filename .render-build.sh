#!/bin/bash
# Build script for Render deployment

set -e

echo "Installing dependencies..."
go mod download

echo "Building application..."
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

echo "Build complete!"

