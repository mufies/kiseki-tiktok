# Stream Service Troubleshooting Guide

## Issue: Browser trying to load RTMP URL instead of HLS URL

### Error Symptoms
```
GET rtmp://localhost:1935/live/sk_xxx net::ERR_UNKNOWN_URL_SCHEME
Autoplay failed: NotSupportedError: Failed to load because no supported source was found.
```

### Root Cause
The browser received an RTMP URL when it expected an HLS URL (HTTP/HTTPS with .m3u8 file).

### Possible Causes

1. **Stream Service Not Running**
   - The stream-service backend might not be running on port 8083
   - Check if the service is running: `curl http://localhost:8083/health`

2. **HLS Transcoding Not Started**
   - RTMP stream connected but HLS transcoding hasn't begun yet
   - Check stream-service logs for FFmpeg process startup

3. **Backend API Returning Wrong URL**
   - The `/streams/{id}/playback` endpoint should return an HLS URL
   - Expected format: `http://localhost:8083/hls/{stream_id}/master.m3u8`
   - NOT: `rtmp://localhost:1935/live/{stream_key}`

### Fixes Applied (Frontend)

1. **URL Validation in API Layer** (`src/api/stream.ts`)
   - Validates playback URL is HTTP/HTTPS before returning
   - Logs backend response for debugging
   - Throws clear error if invalid URL received

2. **StreamPlayer Guards** (`src/components/StreamPlayer.tsx`)
   - Rejects non-HTTP URLs before attempting to load
   - Shows clear error message to user

3. **Page-Level Guards** (`GoLive.tsx`, `WatchStream.tsx`)
   - Only renders StreamPlayer when valid HLS URL is available
   - Shows loading state while waiting for valid URL
   - Provides user feedback during transitions

### How to Diagnose

1. **Open Browser Console**
   - Look for log: `Backend playback response: {...}`
   - Check what URL is being returned

2. **Test Backend Endpoint Directly**
   ```bash
   # Create a stream
   curl -X POST http://localhost:8080/streams \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -d '{"user_id":"USER_ID","title":"Test Stream"}'

   # Get playback URL
   curl http://localhost:8080/streams/STREAM_ID/playback \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

   Expected response:
   ```json
   {
     "playback_url": "http://localhost:8083/hls/STREAM_ID/master.m3u8",
     "protocol": "hls",
     "note": "Stream must be live to access playback"
   }
   ```

3. **Check Stream Service**
   ```bash
   # Check if stream service is running
   curl http://localhost:8083/health

   # Check if HLS files are being generated
   ls /tmp/hls/STREAM_ID/
   # Should see: master.m3u8, variant playlists, .ts segments
   ```

4. **Verify RTMP Connection**
   ```bash
   # Stream with FFmpeg
   ffmpeg -re -i test.mp4 \
     -c:v libx264 -preset veryfast \
     -c:a aac \
     -f flv rtmp://localhost:1935/live/STREAM_KEY

   # Check stream-service logs
   # Should see: "OnPublish received", "Starting HLS transcoder", etc.
   ```

### Required Services

For streaming to work, these services must be running:

1. **API Gateway** (port 8080)
   - Routes requests to microservices

2. **Stream Service** (port 8083, 1935, 50055)
   - Handles RTMP ingestion (1935)
   - Serves HLS files (8083)
   - Provides gRPC API (50055)

3. **Supporting Services**
   - PostgreSQL (stream metadata)
   - Redis (viewer counts)
   - Kafka (events)
   - MinIO (VOD storage)

### Expected Flow

```
1. User creates stream → Backend generates stream_key
2. User starts OBS/FFmpeg → RTMP connection to port 1935
3. Stream service validates stream_key
4. HLS transcoder starts → FFmpeg converts RTMP to HLS
5. HLS segments written to /tmp/hls/{stream_id}/
6. Frontend requests playback URL
7. Backend returns: http://localhost:8083/hls/{stream_id}/master.m3u8
8. Frontend plays HLS stream
```

### Quick Fix Checklist

- [ ] Stream service is running (`docker ps` or check process)
- [ ] RTMP server is accepting connections (port 1935)
- [ ] HLS HTTP server is running (port 8083)
- [ ] Stream status is 'live' (check database or API)
- [ ] HLS files exist in `/tmp/hls/{stream_id}/`
- [ ] Backend `/playback` endpoint returns HTTP URL not RTMP URL
- [ ] API Gateway is routing `/streams/*` to stream service
- [ ] No CORS issues (check browser network tab)

### Common Mistakes

1. **Calling playback URL before streaming starts**
   - Wait for stream.status === 'live'
   - Frontend now polls status every 3 seconds

2. **Expecting immediate HLS availability**
   - HLS transcoding takes 5-10 seconds to start
   - First segments need to be written

3. **Wrong playback URL**
   - Must use `/hls/{stream_id}/master.m3u8` endpoint
   - NOT the RTMP ingest URL

### Testing Without OBS

```bash
# Generate a test video file
ffmpeg -f lavfi -i testsrc=duration=60:size=1280x720:rate=30 \
  -f lavfi -i sine=frequency=1000:duration=60 \
  -pix_fmt yuv420p test.mp4

# Stream the test file
ffmpeg -re -i test.mp4 \
  -c:v libx264 -preset veryfast -maxrate 3000k -bufsize 6000k \
  -c:a aac -b:a 160k \
  -f flv rtmp://localhost:1935/live/YOUR_STREAM_KEY
```

### Need More Help?

Check the stream-service logs:
```bash
# If running in Docker
docker logs stream-service

# If running locally
# Check the terminal where stream-service is running
```

Look for errors related to:
- FFmpeg process startup
- HLS file writing
- RTMP connection handling
