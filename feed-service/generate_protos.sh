#!/bin/bash
# Generate Python gRPC stubs from proto files

set -e

PROTO_DIR="../proto"
OUT_DIR="app/grpc_stubs"

# Activate virtual environment if needed
if [ -d "venv" ]; then
    source venv/bin/activate
fi

echo "Generating gRPC stubs..."

# Generate for each proto file
for proto_file in "$PROTO_DIR"/*.proto; do
    filename=$(basename "$proto_file")
    echo "  - Processing $filename"

    python -m grpc_tools.protoc \
        -I"$PROTO_DIR" \
        --python_out="$OUT_DIR" \
        --grpc_python_out="$OUT_DIR" \
        "$proto_file"
done

echo "Fixing imports in generated files..."

# Fix imports in *_pb2_grpc.py files to use relative imports
for grpc_file in "$OUT_DIR"/*_pb2_grpc.py; do
    if [ -f "$grpc_file" ]; then
        filename=$(basename "$grpc_file")
        echo "  - Fixing imports in $filename"

        # Replace "import xxx_pb2 as" with "from . import xxx_pb2 as"
        sed -i 's/^import \(.*_pb2\) as/from . import \1 as/' "$grpc_file"
    fi
done

echo "Done! gRPC stubs generated successfully."
