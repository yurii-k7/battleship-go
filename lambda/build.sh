#!/bin/bash

set -e

echo "Building Lambda function..."

# Clean previous builds
rm -f bootstrap bootstrap-ws

# Copy backend source to lambda directory
cp -r ../backend/internal ./

# Initialize go module if it doesn't exist
if [ ! -f go.mod ]; then
    go mod init battleship-lambda
fi

# Add dependencies
go mod tidy

# Build the main API handler
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap main.go

# Build WebSocket handler (simplified version)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap-ws websocket.go

# Clean up copied files
rm -rf internal

echo "Lambda functions built successfully!"
echo "Files created:"
echo "  - bootstrap (API handler)"
echo "  - bootstrap-ws (WebSocket handler)"
