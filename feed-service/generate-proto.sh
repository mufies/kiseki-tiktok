#!/bin/bash

# Script to generate Python protobuf files for feed-service
# Usage: ./generate-proto.sh

set -e

PROTO_DIR="../proto"
OUT_DIR="./app/grpc_stubs"

echo "Generating Python protobuf files from $PROTO_DIR..."

# Create output directory if it doesn't exist
mkdir -p "$OUT_DIR"

# Generate Python protobuf and gRPC code for all proto files
for proto_file in "$PROTO_DIR"/*.proto; do
    filename=$(basename "$proto_file")
    echo "  - Generating from $filename..."

    python -m grpc_tools.protoc \
        --proto_path="$PROTO_DIR" \
        --python_out="$OUT_DIR" \
        --grpc_python_out="$OUT_DIR" \
        "$proto_file"
done

# Create __init__.py if it doesn't exist
touch "$OUT_DIR/__init__.py"

echo "✓ Protobuf files generated successfully in $OUT_DIR"
