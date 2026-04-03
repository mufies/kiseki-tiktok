#!/bin/bash

echo "🔍 Searching for the code that constructs the wrong playback URL..."
echo ""

# Search in parent directory for video service
cd ..

echo "Looking for 'rtmp://' URL construction in code..."
grep -r "rtmp://.*1935.*live" --include="*.ts" --include="*.js" --include="*.go" --include="*.java" . 2>/dev/null | head -10

echo ""
echo "Looking for 'playback' related code..."
grep -r "playbackUrl\|playback_url\|playback.*rtmp" --include="*.ts" --include="*.js" --include="*.go" --include="*.java" . 2>/dev/null | head -10

echo ""
echo "Looking for stream service client calls..."
grep -r "getStream\|GetStream" --include="*.ts" --include="*.js" --include="*.go" --include="*.java" . 2>/dev/null | head -10
