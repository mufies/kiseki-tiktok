#!/bin/bash

# Stream Service Diagnostic Script
# Usage: ./diagnose-stream.sh [stream_id]

echo "=========================================="
echo "🔍 Stream Service Diagnostic Tool"
echo "=========================================="
echo ""

STREAM_ID=$1

if [ -z "$STREAM_ID" ]; then
  echo "⚠️  No stream ID provided"
  echo "Usage: ./diagnose-stream.sh <stream_id>"
  echo ""
  echo "Performing general diagnostics..."
  echo ""
else
  echo "📺 Stream ID: $STREAM_ID"
  echo ""
fi

# Check if ports are listening
echo "1️⃣  Checking if services are running..."
echo ""

echo "  🔌 Port 8080 (API Gateway):"
if nc -z localhost 8080 2>/dev/null; then
  echo "     ✅ Listening"
else
  echo "     ❌ Not listening"
fi

echo "  🔌 Port 8083 (Stream Service HTTP):"
if nc -z localhost 8083 2>/dev/null; then
  echo "     ✅ Listening"
else
  echo "     ❌ Not listening"
fi

echo "  🔌 Port 1935 (RTMP Server):"
if nc -z localhost 1935 2>/dev/null; then
  echo "     ✅ Listening"
else
  echo "     ❌ Not listening"
fi

echo ""

# Check FFmpeg
echo "2️⃣  Checking FFmpeg installation..."
echo ""

if command -v ffmpeg &> /dev/null; then
  FFMPEG_VERSION=$(ffmpeg -version | head -n 1)
  echo "  ✅ FFmpeg installed: $FFMPEG_VERSION"
else
  echo "  ❌ FFmpeg not found"
fi

echo ""

# Check HLS directory
echo "3️⃣  Checking HLS directory..."
echo ""

if [ -d "/tmp/hls" ]; then
  echo "  ✅ /tmp/hls exists"

  # Check permissions
  PERMS=$(stat -c %a /tmp/hls 2>/dev/null || stat -f %A /tmp/hls 2>/dev/null)
  echo "  📁 Permissions: $PERMS"

  # Count directories (streams)
  STREAM_COUNT=$(find /tmp/hls -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
  echo "  📊 Active stream directories: $STREAM_COUNT"

  # List streams
  if [ $STREAM_COUNT -gt 0 ]; then
    echo "  📂 Stream directories:"
    find /tmp/hls -mindepth 1 -maxdepth 1 -type d -exec basename {} \; 2>/dev/null | sed 's/^/     /'
  fi
else
  echo "  ❌ /tmp/hls does not exist"
  echo "  💡 Fix: sudo mkdir -p /tmp/hls && sudo chmod 777 /tmp/hls"
fi

echo ""

# Check specific stream if provided
if [ -n "$STREAM_ID" ]; then
  echo "4️⃣  Checking stream $STREAM_ID..."
  echo ""

  if [ -d "/tmp/hls/$STREAM_ID" ]; then
    echo "  ✅ Stream directory exists"

    # Check for master.m3u8
    if [ -f "/tmp/hls/$STREAM_ID/master.m3u8" ]; then
      echo "  ✅ master.m3u8 exists"

      # Count segments
      TS_COUNT=$(find /tmp/hls/$STREAM_ID -name "*.ts" 2>/dev/null | wc -l)
      echo "  📊 Segment files (.ts): $TS_COUNT"
    else
      echo "  ❌ master.m3u8 not found"
    fi

    # List all files
    echo "  📄 Files in directory:"
    ls -lh /tmp/hls/$STREAM_ID/ 2>/dev/null | tail -n +2 | awk '{print "     " $9 " (" $5 ")"}'

  else
    echo "  ❌ Stream directory does not exist: /tmp/hls/$STREAM_ID"
  fi

  echo ""

  # Test HLS URL
  echo "5️⃣  Testing HLS URL..."
  echo ""

  HLS_URL="http://localhost:8083/hls/$STREAM_ID/master.m3u8"
  echo "  🌐 Testing: $HLS_URL"

  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$HLS_URL" 2>/dev/null)

  if [ "$HTTP_CODE" = "200" ]; then
    echo "  ✅ HTTP $HTTP_CODE - HLS file is accessible!"
  elif [ "$HTTP_CODE" = "404" ]; then
    echo "  ❌ HTTP $HTTP_CODE - HLS file not found"
    echo "     💡 The stream might not be transcoding yet"
  else
    echo "  ⚠️  HTTP $HTTP_CODE - Unexpected response"
  fi

  echo ""
fi

# Check running processes
echo "6️⃣  Checking running processes..."
echo ""

FFMPEG_PROCS=$(ps aux | grep -i ffmpeg | grep -v grep | wc -l)
if [ $FFMPEG_PROCS -gt 0 ]; then
  echo "  ✅ FFmpeg processes running: $FFMPEG_PROCS"
  echo "  📋 Processes:"
  ps aux | grep -i ffmpeg | grep -v grep | awk '{print "     PID " $2 ": " $11}' | head -5
else
  echo "  ⚠️  No FFmpeg processes found"
  echo "     💡 This might be normal if no streams are active"
fi

echo ""

# Summary
echo "=========================================="
echo "📊 Summary"
echo "=========================================="
echo ""

ISSUES=0

if ! nc -z localhost 8083 2>/dev/null; then
  echo "❌ Stream service (port 8083) is not running"
  ISSUES=$((ISSUES+1))
fi

if ! nc -z localhost 1935 2>/dev/null; then
  echo "❌ RTMP server (port 1935) is not running"
  ISSUES=$((ISSUES+1))
fi

if ! command -v ffmpeg &> /dev/null; then
  echo "❌ FFmpeg is not installed"
  ISSUES=$((ISSUES+1))
fi

if [ ! -d "/tmp/hls" ]; then
  echo "❌ HLS directory (/tmp/hls) does not exist"
  ISSUES=$((ISSUES+1))
fi

if [ -n "$STREAM_ID" ] && [ ! -d "/tmp/hls/$STREAM_ID" ]; then
  echo "❌ Stream directory does not exist"
  ISSUES=$((ISSUES+1))
fi

if [ -n "$STREAM_ID" ] && [ "$HTTP_CODE" != "200" ]; then
  echo "❌ HLS URL is not accessible (HTTP $HTTP_CODE)"
  ISSUES=$((ISSUES+1))
fi

if [ $ISSUES -eq 0 ]; then
  echo "✅ All checks passed!"
else
  echo ""
  echo "Found $ISSUES issue(s). See above for details."
fi

echo ""
echo "=========================================="
echo ""

# Provide next steps
if [ -n "$STREAM_ID" ] && [ "$HTTP_CODE" = "404" ]; then
  echo "💡 Next Steps:"
  echo "   1. Make sure OBS is connected and streaming"
  echo "   2. Check stream service logs for errors"
  echo "   3. Verify stream status is 'live' in database"
  echo "   4. Wait 10-15 seconds for transcoding to start"
  echo "   5. Run this script again to verify"
  echo ""
fi
