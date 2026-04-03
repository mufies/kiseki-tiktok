#!/bin/bash

# API Usage Examples for Stream Service with HLS

BASE_URL="http://localhost:8083"

echo "🎥 Stream Service API Examples"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 1. Create a new stream
echo -e "${BLUE}1. Create a new stream${NC}"
echo "POST /streams"
echo ""

STREAM_RESPONSE=$(curl -s -X POST "$BASE_URL/streams" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "My Live Stream",
    "description": "Testing HLS streaming"
  }')

echo "$STREAM_RESPONSE" | jq .
STREAM_ID=$(echo "$STREAM_RESPONSE" | jq -r '.id')
STREAM_KEY=$(echo "$STREAM_RESPONSE" | jq -r '.stream_key')

echo ""
echo -e "${GREEN}✅ Stream created!${NC}"
echo -e "Stream ID: ${YELLOW}$STREAM_ID${NC}"
echo -e "Stream Key: ${YELLOW}$STREAM_KEY${NC}"
echo ""
read -p "Press Enter to continue..."
echo ""

# 2. Get stream details
echo -e "${BLUE}2. Get stream details${NC}"
echo "GET /streams/$STREAM_ID"
echo ""

curl -s "$BASE_URL/streams/$STREAM_ID" | jq .

echo ""
read -p "Press Enter to continue..."
echo ""

# 3. Instructions for OBS
echo -e "${BLUE}3. Configure OBS Studio${NC}"
echo "================================"
echo ""
echo "Server:"
echo -e "  ${GREEN}rtmp://localhost:1935/live${NC}"
echo ""
echo "Stream Key:"
echo -e "  ${GREEN}$STREAM_KEY${NC}"
echo ""
echo "🎬 Start streaming from OBS now!"
echo ""
read -p "Press Enter once you've started streaming from OBS..."
echo ""

# 4. Start stream (optional, usually auto-started by RTMP)
echo -e "${BLUE}4. Start stream${NC}"
echo "POST /streams/$STREAM_ID/start"
echo ""

curl -s -X POST "$BASE_URL/streams/$STREAM_ID/start" | jq .

echo ""
read -p "Press Enter to continue..."
echo ""

# Wait for HLS segments to be generated
echo -e "${YELLOW}⏳ Waiting for HLS segments to be generated (10 seconds)...${NC}"
sleep 10
echo ""

# 5. Get playback URL
echo -e "${BLUE}5. Get HLS playback URL${NC}"
echo "GET /streams/$STREAM_ID/playback"
echo ""

PLAYBACK_RESPONSE=$(curl -s "$BASE_URL/streams/$STREAM_ID/playback")
echo "$PLAYBACK_RESPONSE" | jq .

HLS_URL=$(echo "$PLAYBACK_RESPONSE" | jq -r '.hls_url')
FULL_HLS_URL="$BASE_URL$HLS_URL"

echo ""
echo -e "${GREEN}✅ HLS Playlist URL:${NC}"
echo -e "  ${YELLOW}$FULL_HLS_URL${NC}"
echo ""

# 6. Test HLS playback
echo -e "${BLUE}6. Test HLS playback${NC}"
echo ""
echo "Option 1 - Browser (Video.js):"
echo -e "  ${GREEN}file://$(pwd)/../test_player.html?streamId=$STREAM_ID${NC}"
echo ""
echo "Option 2 - ffplay (command line):"
echo -e "  ${GREEN}ffplay $FULL_HLS_URL${NC}"
echo ""
echo "Option 3 - VLC:"
echo -e "  ${GREEN}vlc $FULL_HLS_URL${NC}"
echo ""
echo "Option 4 - curl (check playlist):"
echo -e "  ${GREEN}curl $FULL_HLS_URL${NC}"
echo ""

read -p "Press Enter to test with curl..."
echo ""

curl -s "$FULL_HLS_URL"
echo ""
echo ""

# 7. Get live streams
echo -e "${BLUE}7. Get all live streams${NC}"
echo "GET /streams/live"
echo ""

curl -s "$BASE_URL/streams/live" | jq .

echo ""
read -p "Press Enter to continue..."
echo ""

# 8. Join stream (increment viewer count)
echo -e "${BLUE}8. Join stream (as viewer)${NC}"
echo "POST /streams/$STREAM_ID/viewers/join"
echo ""

curl -s -X POST "$BASE_URL/streams/$STREAM_ID/viewers/join" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }' | jq .

echo ""
read -p "Press Enter to continue..."
echo ""

# 9. Get updated stream info (should show viewer count)
echo -e "${BLUE}9. Get updated stream info${NC}"
echo "GET /streams/$STREAM_ID"
echo ""

curl -s "$BASE_URL/streams/$STREAM_ID" | jq .

echo ""
echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✅ API Examples Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "🎥 Your stream is live at:"
echo -e "  ${YELLOW}$FULL_HLS_URL${NC}"
echo ""
echo "📝 To end the stream:"
echo "  1. Stop streaming in OBS"
echo "  2. Or call: curl -X POST $BASE_URL/streams/$STREAM_ID/end"
echo ""
echo "🧹 To delete the stream:"
echo "  curl -X DELETE $BASE_URL/streams/$STREAM_ID"
echo ""
