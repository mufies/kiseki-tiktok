#!/bin/sh

set -e

PROTO_DIR="../proto"
OUT_DIR="./internal/grpc/interactionpb"

echo "Generating interaction protobuf files..."

mkdir -p "$OUT_DIR"

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/interaction.proto"

echo "✓ Interaction protobuf files generated successfully in $OUT_DIR"
