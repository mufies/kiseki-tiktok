#!/bin/bash

echo "🔄 Restarting Stream Service"
echo "================================"
echo ""

echo "1️⃣  Stopping current container..."
docker stop tiktok-stream-service

echo ""
echo "2️⃣  Removing old container..."
docker rm tiktok-stream-service

echo ""
echo "3️⃣  Rebuilding with latest code..."
docker-compose build stream-service

echo ""
echo "4️⃣  Starting updated service..."
docker-compose up -d stream-service

echo ""
echo "5️⃣  Waiting for service to start..."
sleep 5

echo ""
echo "6️⃣  Testing health endpoint..."
curl -s http://localhost:8083/health | jq .

echo ""
echo "================================"
echo "✅ Service Restarted!"
echo "================================"

echo ""
echo "Now testing playback URL with test stream..."
STREAM_ID="4b80cf48-030a-4bcc-917f-be3382f89931"
echo "Stream ID: $STREAM_ID"
curl -s "http://localhost:8083/streams/$STREAM_ID/playback"

echo ""
