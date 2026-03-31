#!/bin/bash
# Fix protobuf version mismatch in feed service

set -e

echo "Regenerating proto files with compatible version..."

cd /home/mufies/Code/tiktok-clone/feed-service

# Remove old generated files
rm -rf app/grpc_stubs/*.py
rm -rf app/grpc_stubs/__pycache__

# Ensure grpc_stubs directory exists
mkdir -p app/grpc_stubs

# Generate proto files
python3 -m grpc_tools.protoc \
    --proto_path=../proto \
    --python_out=app/grpc_stubs \
    --grpc_python_out=app/grpc_stubs \
    --pyi_out=app/grpc_stubs \
    ../proto/video.proto \
    ../proto/event.proto \
    ../proto/interaction.proto \
    ../proto/user.proto \
    ../proto/notification.proto

# Create __init__.py
touch app/grpc_stubs/__init__.py

echo "Proto files regenerated successfully!"
ls -la app/grpc_stubs/
