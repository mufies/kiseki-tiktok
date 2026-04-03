#!/bin/bash

# Find Stream ID by Stream Key
# Usage: ./find_stream_id.sh sk_9eb675c255f14fa38e60a89ca5da1420

STREAM_KEY="${1:-sk_9eb675c255f14fa38e60a89ca5da1420}"

echo "Looking up stream for key: $STREAM_KEY"
echo ""

# Database connection details (adjust if needed)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-stream_service}"

# Query the database
RESULT=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
  "SELECT id, title, status FROM streams WHERE stream_key = '$STREAM_KEY';")

if [ -z "$RESULT" ]; then
  echo "❌ No stream found with key: $STREAM_KEY"
  echo ""
  echo "Create a stream first:"
  echo "  curl -X POST http://localhost:8083/streams \\"
  echo "    -H 'Content-Type: application/json' \\"
  echo "    -d '{\"user_id\":\"YOUR_USER_UUID\",\"title\":\"Test Stream\",\"description\":\"Testing\"}'"
  exit 1
fi

STREAM_ID=$(echo "$RESULT" | awk '{print $1}' | xargs)
TITLE=$(echo "$RESULT" | awk '{print $3}' | xargs)
STATUS=$(echo "$RESULT" | awk '{print $5}' | xargs)

echo "✅ Stream found!"
echo "   Stream ID: $STREAM_ID"
echo "   Title: $TITLE"
echo "   Status: $STATUS"
echo ""
echo "📺 HLS Playback URL (use this in VLC):"
echo "   http://localhost:8083/hls/$STREAM_ID/master.m3u8"
echo ""
echo "📡 RTMP Publish URL (use this in OBS):"
echo "   Server: rtmp://localhost:1935/live"
echo "   Stream Key: $STREAM_KEY"
echo ""
echo "🎬 To play in VLC:"
echo "   vlc http://localhost:8083/hls/$STREAM_ID/master.m3u8"
