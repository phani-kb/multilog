#!/bin/bash

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build the project
echo "Building the project..."
go build ./...

echo "Build completed successfully."
