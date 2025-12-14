#!/bin/bash
# Run script for ROCm Monitor

# Change to the script's directory
cd "$(dirname "$0")"

# Show current directory
echo "Running from: $(pwd)"

# Run the Go application
go run . "$@"