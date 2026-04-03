#!/bin/bash

# HLS Transcoding Test Script
# This script helps you quickly test the HLS streaming functionality

set -e

echo "🎥 HLS Streaming Test Script"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if FFmpeg is installed
echo "1️⃣  Checking FFmpeg installation..."
if command -v ffmpeg &> /dev/null; then
    echo -e "${GREEN}✅ FFmpeg is installed${NC}"
    ffmpeg -version | head -n 1
else
    echo -e "${RED}❌ FFmpeg is NOT installed${NC}"
    echo ""
    echo "Please install FFmpeg:"
    echo "  Ubuntu/Debian: sudo apt install ffmpeg"
    echo "  macOS: brew install ffmpeg"
    echo "  Arch: sudo pacman -S ffmpeg"
    exit 1
fi
echo ""

# Check if service is running
echo "2️⃣  Checking if stream service is running..."
if curl -s http://localhost:8083/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Stream service is running on port 8083${NC}"
else
    echo -e "${RED}❌ Stream service is NOT running${NC}"
    echo ""
    echo "Please start the service first:"
    echo "  go run cmd/main.go"
    exit 1
fi
echo ""

# Check RTMP port
echo "3️⃣  Checking RTMP server..."
if netstat -tuln 2>/dev/null | grep -q ":1935 " || ss -tuln 2>/dev/null | grep -q ":1935 "; then
    echo -e "${GREEN}✅ RTMP server is listening on port 1935${NC}"
else
    echo -e "${YELLOW}⚠️  Cannot detect RTMP on port 1935${NC}"
    echo "   (This might be a false negative if using Docker)"
fi
echo ""

# Check /tmp/hls directory
echo "4️⃣  Checking HLS output directory..."
if [ -d "/tmp/hls" ]; then
    echo -e "${GREEN}✅ /tmp/hls directory exists${NC}"
    echo "   Current streams:"
    ls -1 /tmp/hls/ 2>/dev/null | while read stream; do
        if [ -f "/tmp/hls/$stream/playlist.m3u8" ]; then
            echo -e "   ${GREEN}▶️  $stream (ACTIVE)${NC}"
        else
            echo -e "   ${YELLOW}⏸️  $stream (no playlist)${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️  /tmp/hls directory does not exist yet${NC}"
    echo "   It will be created when the first stream starts"
fi
echo ""

# Instructions
echo "================================"
echo "📝 Next Steps:"
echo "================================"
echo ""
echo "1. Create a stream in the database (or use API)"
echo "   Example SQL:"
echo "   INSERT INTO streams (id, user_id, title, stream_key, status)"
echo "   VALUES (gen_random_uuid(), 'your-user-id', 'Test Stream', 'test123', 'scheduled');"
echo ""
echo "2. Configure OBS Studio:"
echo "   Server: rtmp://localhost:1935/live"
echo "   Stream Key: test123  (or your stream_key from database)"
echo ""
echo "3. Start streaming from OBS"
echo ""
echo "4. Open test player:"
echo "   file://$(pwd)/test_player.html"
echo ""
echo "5. Enter your stream ID and click 'Load Stream'"
echo ""
echo "================================"
echo "🔍 Useful Commands:"
echo "================================"
echo ""
echo "# Watch FFmpeg processes"
echo "watch -n 1 'ps aux | grep ffmpeg'"
echo ""
echo "# Monitor HLS directory"
echo "watch -n 1 'ls -lh /tmp/hls/*/'"
echo ""
echo "# Test HLS playback with ffplay"
echo "ffplay http://localhost:8083/hls/{stream_id}/playlist.m3u8"
echo ""
echo "# Check service logs"
echo "# (Look for [RTMP] and [Transcoder] messages)"
echo ""
echo "================================"
echo -e "${GREEN}✅ All checks passed!${NC}"
echo "================================"
