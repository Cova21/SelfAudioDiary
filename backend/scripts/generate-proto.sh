#!/bin/bash

# Script to generate Go code from protobuf files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$BACKEND_DIR/internal/protobuf"
GEN_DIR="$BACKEND_DIR/internal/gen"

echo "Backend dir: $BACKEND_DIR"
echo "Proto dir: $PROTO_DIR"
echo "Gen dir: $GEN_DIR"

# Create gen directory if it doesn't exist
mkdir -p "$GEN_DIR"

# Install protoc plugins if not already installed
echo "Installing protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "protoc is not installed. Please install Protocol Buffers compiler."
    echo "Visit: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "Generating Go code from proto files..."

# Generate code for each proto file
for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        filename=$(basename "$proto_file")
        echo "Processing $filename..."
        
        protoc \
            --proto_path="$PROTO_DIR" \
            --go_out="$GEN_DIR" \
            --go_opt=paths=source_relative \
            --go-grpc_out="$GEN_DIR" \
            --go-grpc_opt=paths=source_relative \
            "$proto_file"
    fi
done

echo "Proto generation completed successfully!"
echo "Generated files are in: $GEN_DIR"
