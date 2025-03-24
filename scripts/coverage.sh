#!/bin/bash

# Create coverage folder if it doesn't exist
mkdir -p coverage

# Run tests with coverage
go test -coverprofile=coverage/coverage.out ./...

# Generate HTML report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

echo "Coverage report generated at coverage/coverage.html"
