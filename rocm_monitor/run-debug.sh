#!/bin/bash
# Debug run script for ROCm Monitor

# Build with debug symbols
echo "Building with debug symbols..."
go build -gcflags="all=-N -l" -o rocm_monitor_debug

# Check if dlv is installed
if ! command -v dlv &> /dev/null; then
    echo "Delve debugger not installed. Installing..."
    go install github.com/go-delve/delve/cmd/dlv@latest
fi

# Run with debugger
echo "Starting debugger..."
echo "Available commands:"
echo "  - continue (c): Run until breakpoint"
echo "  - break (b) main.main: Set breakpoint at main function"
echo "  - next (n): Step over"
echo "  - step (s): Step into"
echo "  - print (p) <variable>: Print variable value"
echo ""

dlv exec ./rocm_monitor_debug -- "$@"