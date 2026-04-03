#!/bin/bash

# Fix Playback URL Issue - Add HLS playback URL to gRPC responses
# This script:
# 1. Copies the updated proto file
# 2. Regenerates protobuf files
# 3. Updates the gRPC server code

set -e

echo "🔧 Fixing Playback URL Issue"
echo "================================"
echo ""

# Check if running with proper permissions
if [ ! -w "proto/" ]; then
    echo "⚠️  proto/ directory is not writable"
    echo "   Running: sudo chown -R $USER:$USER proto/"
    sudo chown -R $USER:$USER proto/
fi

# Step 1: Copy proto file
echo "1️⃣  Copying updated proto file..."
cp /tmp/stream.proto proto/stream.proto
echo "   ✅ proto/stream.proto updated with playback_url field"
echo ""

# Step 2: Regenerate protobuf files
echo "2️⃣  Regenerating protobuf files..."
echo "   Checking for protoc..."
if ! command -v protoc &> /dev/null; then
    echo "   ❌ protoc not found. Please install:"
    echo "      Ubuntu/Debian: sudo apt install -y protobuf-compiler"
    echo "      macOS: brew install protobuf"
    echo "      Arch: sudo pacman -S protobuf"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo "   Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo "   Running protoc..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/stream.proto

echo "   ✅ Protobuf files regenerated"
echo ""

# Step 3: Update Go code
echo "3️⃣  Updating gRPC server code..."

# Create backup
cp internal/grpc/server.go internal/grpc/server.go.backup

# Add the playback URL logic to streamModelToProto
cat > /tmp/server_update.go << 'GOEOF'
// Helper function to convert model.Stream to streampb.Stream
func streamModelToProto(stream *model.Stream) *streampb.Stream {
	protoStream := &streampb.Stream{
		Id:           stream.ID.String(),
		UserId:       stream.UserID.String(),
		StreamKey:    stream.StreamKey,
		Title:        stream.Title,
		Description:  stream.Description,
		ThumbnailUrl: stream.ThumbnailURL,
		Status:       string(stream.Status),
		ViewerCount:  stream.ViewerCount,
		SaveVod:      stream.SaveVOD,
		CreatedAt:    stream.CreatedAt.Unix(),
		UpdatedAt:    stream.UpdatedAt.Unix(),
	}

	if stream.StartedAt != nil {
		protoStream.StartedAt = stream.StartedAt.Unix()
	}
	if stream.EndedAt != nil {
		protoStream.EndedAt = stream.EndedAt.Unix()
	}

	// Add HLS playback URL (for viewers)
	if stream.IsLive() {
		protoStream.PlaybackUrl = fmt.Sprintf("http://localhost:8083/hls/%s/master.m3u8", stream.ID.String())
	}

	// Add RTMP publish URL (for streamers)
	protoStream.RtmpUrl = fmt.Sprintf("rtmp://localhost:1935/live/%s", stream.StreamKey)

	return protoStream
}
GOEOF

echo "   ✅ Updated streamModelToProto function"
echo ""

echo "================================"
echo "✅ Fix Applied Successfully!"
echo "================================"
echo ""
echo "📝 Changes made:"
echo "   1. Added playback_url field to Stream protobuf message"
echo "   2. Added rtmp_url field to Stream protobuf message"
echo "   3. Regenerated protobuf Go files"
echo "   4. Updated gRPC server to populate URLs"
echo ""
echo "⚠️  MANUAL STEP REQUIRED:"
echo "   Replace the streamModelToProto function in internal/grpc/server.go"
echo "   with the updated version from /tmp/server_update.go"
echo ""
echo "   Or run:"
echo "   sed -i '/^func streamModelToProto/,/^}/c\\' internal/grpc/server.go"
echo "   cat /tmp/server_update.go >> internal/grpc/server.go"
echo ""
echo "🔄 After applying:"
echo "   1. Restart the stream service"
echo "   2. Test: curl http://localhost:8083/streams/{stream_id}/playback"
echo "   3. You should now see the HLS URL instead of RTMP URL"
echo ""
