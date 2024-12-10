#!/bin/bash

set -e

# Get the directory containing this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Build the binary
echo "Building take binary..."
cd "$PROJECT_ROOT"
go build -o bin/take ./cmd/take

# Make it executable
chmod +x bin/take

echo "Build complete. Binary is at bin/take"