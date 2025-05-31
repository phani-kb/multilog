#!/bin/bash

# Create coverage folder if it doesn't exist
mkdir -p coverage

# Run tests with coverage, excluding examples and cmd directories
go test -coverprofile=coverage/coverage.out $(go list ./... | grep -v "/examples\|/cmd")

# Generate HTML report
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Extract coverage percentage and check if it meets the threshold
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
THRESHOLD=80

echo "Coverage report generated at coverage/coverage.html"
echo "Total coverage: $COVERAGE%"

# Check if coverage is below the threshold
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
  echo "FAILURE: Code coverage is below the minimum threshold of $THRESHOLD%"
  exit 1
fi
