#!/bin/sh

# Script to generate protobuf files for video-service
# Usage: ./generate-proto.sh

set -e

PROTO_DIR="../proto"
OUT_DIR="./internal/grpc/videopb"

echo "Generating protobuf files from $PROTO_DIR/video.proto..."

# Create output directory if it doesn't exist
mkdir -p "$OUT_DIR"

# Generate Go protobuf and gRPC code
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/video.proto"

echo "✓ Protobuf files generated successfully in $OUT_DIR"
