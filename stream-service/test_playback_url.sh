#!/bin/bash

# Test script to get the correct HLS playback URL

STREAM_ID="${1}"

if [ -z "$STREAM_ID" ]; then
  echo "Usage: ./test_playback_url.sh <stream_id>"
  echo "Example: ./test_playback_url.sh 550e8400-e29b-41d4-a716-446655440000"
  exit 1
fi

echo "Getting playback URL for stream: $STREAM_ID"
echo ""

# Call the playback endpoint
RESPONSE=$(curl -s "http://localhost:8083/streams/$STREAM_ID/playback")

echo "Response:"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"

# Extract the playback URL
PLAYBACK_URL=$(echo "$RESPONSE" | jq -r '.playback_url' 2>/dev/null)

if [ "$PLAYBACK_URL" != "null" ] && [ -n "$PLAYBACK_URL" ]; then
  echo ""
  echo "✅ HLS Playback URL:"
  echo "   $PLAYBACK_URL"
  echo ""
  echo "🎬 Play in VLC:"
  echo "   vlc \"$PLAYBACK_URL\""
else
  echo ""
  echo "❌ Failed to get playback URL. Make sure:"
  echo "   1. Stream service is running"
  echo "   2. Stream ID is correct"
  echo "   3. Stream is live (actively streaming from OBS)"
fi
