# Streaming Setup Guide

## Overview

This project uses **MediaMTX** to handle RTMP to HLS transcoding for live streaming.

## Architecture

```
OBS/FFmpeg --> RTMP (port 1935) --> MediaMTX --> HLS (port 8888) --> Browser
                                         |
                                         v
                                  Stream Service validates stream keys
```

## Setup Instructions

### 1. Start MediaMTX Server

```bash
cd /home/mufies/Code/tiktok-clone
./mediamtx mediamtx-custom.yml
```

MediaMTX will start and listen on:
- **RTMP**: Port 1935 (for OBS input)
- **HLS**: Port 8888 (for browser playback)

### 2. Configure Frontend Environment

Create `.env` file in `tiktok-test-api/`:

```bash
cp .env.example .env
```

Make sure it contains:
```
VITE_API_URL=http://localhost:8080
VITE_RTMP_URL=rtmp://localhost:1935
```

### 3. Start Services

```bash
# Start stream service
cd stream-service
go run cmd/main.go

# Start frontend
cd ../tiktok-test-api
npm run dev
```

### 4. Go Live

1. Open frontend at http://localhost:5173
2. Login and click "Go Live"
3. Fill in stream title and generate stream key
4. Copy the RTMP URL and Stream Key

### 5. Configure OBS

1. Open OBS Studio
2. Go to Settings → Stream
3. Set:
   - **Service**: Custom
   - **Server**: `rtmp://localhost:1935`
   - **Stream Key**: `{your_generated_stream_key}`
4. Click "Start Streaming"

### 6. Watch Stream

After OBS connects:
1. The frontend will detect the stream is live (status updates every 3s)
2. Click "Start Broadcasting"
3. Your stream will be available at:
   - **HLS URL**: `http://localhost:8888/{stream_key}/index.m3u8`

## Troubleshooting

### Stream not showing "Connected"
- Check MediaMTX is running: `ps aux | grep mediamtx`
- Check OBS connection status
- View MediaMTX logs for errors

### Video player not loading
- Open browser console for errors
- Check HLS URL is accessible: `curl http://localhost:8888/{stream_key}/index.m3u8`
- Verify stream status is "live" in API response

### CORS errors
- MediaMTX config has `hlsAllowOrigins: ['*']` to allow all origins
- Check browser console for specific CORS errors

## MediaMTX Configuration

Config file: `/home/mufies/Code/tiktok-clone/mediamtx-custom.yml`

Key settings:
- **RTMP Port**: 1935
- **HLS Port**: 8888
- **HLS Variant**: lowLatency (for better performance)
- **Segment Duration**: 1s
- **Part Duration**: 200ms

## Testing with FFmpeg

Instead of OBS, you can test with FFmpeg:

```bash
ffmpeg -re -i video.mp4 \
  -c:v libx264 -preset veryfast \
  -maxrate 3000k -bufsize 6000k \
  -pix_fmt yuv420p -g 50 \
  -c:a aac -b:a 160k -ac 2 -ar 44100 \
  -f flv rtmp://localhost:1935/{your_stream_key}
```

Replace `{your_stream_key}` with your actual stream key from the frontend.
