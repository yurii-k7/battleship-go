#!/bin/bash

set -e

STAGE=${1:-dev}

echo "Deploying to stage: $STAGE"

# Check if serverless is installed
if ! command -v serverless &> /dev/null; then
    echo "Serverless Framework not found. Installing..."
    npm install -g serverless
fi

# Check if required plugins are installed
if [ ! -d "node_modules" ]; then
    echo "Installing Serverless plugins..."
    npm init -y
    npm install serverless-domain-manager
fi

# Build the Lambda functions
echo "Building Lambda functions..."
./build.sh

# Deploy using Serverless Framework
echo "Deploying with Serverless Framework..."
serverless deploy --stage $STAGE

echo "Deployment complete!"
echo ""
echo "API Gateway URL:"
serverless info --stage $STAGE | grep "endpoint:"

echo ""
echo "WebSocket URL:"
serverless info --stage $STAGE | grep "websocket:"
