# Quick Start: ABR HLS Streaming

## 1. Start the Service

```bash
cd /home/mufies/Code/tiktok-clone/stream-service
./bin/stream-service
```

Expected output:
```
[Transcoder] Transcoder manager initialized
[RTMP] Starting RTMP server on port 1935
[HTTP] Stream HTTP server starting on :8083
```

## 2. Create a Stream

```bash
curl -X POST http://localhost:8083/streams \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "My Live Stream",
    "description": "Testing ABR"
  }'
```

Save the returned `id` and `stream_key`.

## 3. Configure OBS

1. **Settings → Stream**
   - Service: Custom
   - Server: `rtmp://localhost:1935/live`
   - Stream Key: `{your_stream_key}`

2. **Settings → Output**
   - Bitrate: 6000 kbps
   - Keyframe Interval: 2

3. **Settings → Video**
   - Resolution: 1920x1080
   - FPS: 30

## 4. Start Streaming

Click "Start Streaming" in OBS.

## 5. Watch the Stream

### Web Player
```bash
# Open in browser
xdg-open /home/mufies/Code/tiktok-clone/stream-service/test_player.html

# Or manually open and enter stream ID
```

### VLC
```bash
vlc http://localhost:8083/hls/{stream_id}/master.m3u8
```

### ffplay
```bash
ffplay http://localhost:8083/hls/{stream_id}/master.m3u8
```

## 6. Verify ABR

Check files:
```bash
ls -la /tmp/hls/{stream_id}/

# Should see:
# master.m3u8
# 1080p/playlist.m3u8
# 720p/playlist.m3u8
# 480p/playlist.m3u8
# 360p/playlist.m3u8
```

View master playlist:
```bash
cat /tmp/hls/{stream_id}/master.m3u8
```

## Troubleshooting

### No HLS files generated
- Wait 6-12 seconds for first segments
- Check FFmpeg is installed: `which ffmpeg`
- Check logs for errors

### Only one quality available
- Verify `ENABLE_ABR=true` in config
- Check logs: `[Transcoder] ABR enabled: true`
- Ensure input resolution is 1080p

### High CPU usage
Enable hardware acceleration:
```bash
export HW_ACCEL_ENABLED=true
export HW_ACCEL_TYPE=nvenc  # or qsv, videotoolbox
```

## API Reference

### Get Playback URL
```bash
curl http://localhost:8083/streams/{stream_id}/playback
```

### Get Stream Info
```bash
curl http://localhost:8083/streams/{stream_id}
```

### End Stream
```bash
curl -X POST http://localhost:8083/streams/{stream_id}/end
```

## Configuration

Edit `.env` or set environment variables:

```bash
# ABR Settings
ENABLE_ABR=true                    # Enable ABR (default: true)
HW_ACCEL_ENABLED=false             # Hardware acceleration (default: false)
HW_ACCEL_TYPE=nvenc                # nvenc, qsv, videotoolbox

# HLS Settings
HLS_SEGMENT_TIME=6                 # Segment duration in seconds
HLS_PLAYLIST_LENGTH=5              # Number of segments to keep

# Server Settings
SERVER_PORT=8083                   # HTTP port
RTMP_PORT=1935                     # RTMP port
GRPC_PORT=50055                    # gRPC port
```

## More Information

- Full testing guide: `ABR_TESTING_GUIDE.md`
- Implementation details: `ABR_IMPLEMENTATION_SUMMARY.md`
- Original documentation: `README_HLS.md`
