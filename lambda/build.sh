#!/bin/bash

set -e

echo "Building Lambda function..."

# Clean previous builds
rm -f bootstrap bootstrap-ws

# Copy backend source to lambda directory
cp -r ../backend/internal ./
cp ../backend/go.mod ./
cp ../backend/go.sum ./

# Build the main API handler
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap main.go

# Build WebSocket handler (simplified version)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap-ws websocket.go

# Clean up copied files
rm -rf internal go.mod go.sum

echo "Lambda functions built successfully!"
echo "Files created:"
echo "  - bootstrap (API handler)"
echo "  - bootstrap-ws (WebSocket handler)"
