# ABR (Adaptive Bitrate) HLS Streaming - Testing Guide

## Overview

The stream-service now supports multi-bitrate Adaptive Bitrate (ABR) HLS streaming. This allows players to automatically switch between different quality levels (1080p, 720p, 480p, 360p) based on available bandwidth.

## Architecture

```
RTMP Input --> FFmpeg (filter_complex + multi-output) --> HLS Variants
                                                              |
                                                              v
                                                    /tmp/hls/{stream_id}/
                                                    +-- master.m3u8
                                                    +-- 1080p/playlist.m3u8
                                                    +-- 720p/playlist.m3u8
                                                    +-- 480p/playlist.m3u8
                                                    +-- 360p/playlist.m3u8
```

## Configuration

The following environment variables can be used to configure ABR:

```bash
# Enable/disable ABR (default: true)
ENABLE_ABR=true

# Enable hardware acceleration (default: false)
HW_ACCEL_ENABLED=false

# Hardware acceleration type: nvenc (NVIDIA), qsv (Intel), videotoolbox (macOS)
HW_ACCEL_TYPE=nvenc

# HLS segment settings
HLS_SEGMENT_TIME=6        # Seconds per segment
HLS_PLAYLIST_LENGTH=5     # Number of segments in playlist
```

## Default Quality Variants

The service generates 4 quality variants by default:

| Quality | Resolution | Video Bitrate | Audio Bitrate | Total Bandwidth |
|---------|------------|---------------|---------------|-----------------|
| 1080p   | 1920x1080  | 5000 kbps     | 128 kbps      | 5128 kbps       |
| 720p    | 1280x720   | 2500 kbps     | 128 kbps      | 2628 kbps       |
| 480p    | 854x480    | 1000 kbps     | 96 kbps       | 1096 kbps       |
| 360p    | 640x360    | 600 kbps      | 64 kbps       | 664 kbps        |

## Testing Steps

### 1. Build and Start the Service

```bash
cd /home/mufies/Code/tiktok-clone/stream-service

# Build
go build -o bin/stream-service ./cmd/main.go

# Start the service
./bin/stream-service
```

The service will start:
- RTMP server on port **1935**
- HTTP server on port **8083**
- gRPC server on port **50055**

### 2. Create a Stream

Use the API to create a stream:

```bash
curl -X POST http://localhost:8083/streams \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "ABR Test Stream",
    "description": "Testing adaptive bitrate streaming"
  }'
```

Response will contain:
- `id`: Stream ID (UUID)
- `stream_key`: RTMP stream key

### 3. Configure OBS Studio

1. **Settings → Stream**
   - Service: Custom
   - Server: `rtmp://localhost:1935/live`
   - Stream Key: `{your_stream_key_from_api}`

2. **Settings → Output**
   - Output Mode: Advanced
   - Encoder: x264 (or hardware encoder)
   - Rate Control: CBR
   - Bitrate: 6000 kbps (higher than highest variant)
   - Keyframe Interval: 2 (IMPORTANT for ABR sync)

3. **Settings → Video**
   - Base Resolution: 1920x1080
   - Output Resolution: 1920x1080
   - FPS: 30

### 4. Start Streaming

1. Click "Start Streaming" in OBS
2. Check service logs for transcoder startup:
   ```
   [Transcoder] Starting FFmpeg for stream {stream_id}
   [Transcoder] ABR enabled: true
   [Transcoder] Variants: 4
   [Transcoder]   - 1080p: 1920x1080 @ 5000k video, 128k audio
   [Transcoder]   - 720p: 1280x720 @ 2500k video, 128k audio
   [Transcoder]   - 480p: 854x480 @ 1000k video, 96k audio
   [Transcoder]   - 360p: 640x360 @ 600k video, 64k audio
   ```

### 5. Verify File Structure

After 6-12 seconds, check the output directory:

```bash
ls -la /tmp/hls/{stream_id}/

# Expected structure:
# master.m3u8         <- Master playlist
# 1080p/
#   ├── playlist.m3u8
#   ├── segment_000.ts
#   ├── segment_001.ts
#   └── ...
# 720p/
#   ├── playlist.m3u8
#   └── segment_*.ts
# 480p/
#   ├── playlist.m3u8
#   └── segment_*.ts
# 360p/
#   ├── playlist.m3u8
#   └── segment_*.ts
```

### 6. View Master Playlist

```bash
cat /tmp/hls/{stream_id}/master.m3u8
```

Expected content:
```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=5128000,RESOLUTION=1920x1080,NAME="1080p"
1080p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2628000,RESOLUTION=1280x720,NAME="720p"
720p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1096000,RESOLUTION=854x480,NAME="480p"
480p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=664000,RESOLUTION=640x360,NAME="360p"
360p/playlist.m3u8
```

### 7. Test Playback

#### Option A: Web Player

1. Open `test_player.html` in a browser
2. Enter your stream ID
3. Click "Load Stream"
4. Click "Toggle Stats" to see quality information

The player will automatically:
- Detect available quality variants
- Switch between qualities based on bandwidth
- Display current quality in stats

#### Option B: VLC Player

```bash
vlc http://localhost:8083/hls/{stream_id}/master.m3u8
```

#### Option C: ffplay

```bash
ffplay http://localhost:8083/hls/{stream_id}/master.m3u8
```

### 8. Test Quality Switching

To verify ABR is working:

1. **Bandwidth Throttling** (Chrome DevTools):
   - Open DevTools (F12)
   - Network tab → Throttling
   - Select "Fast 3G" or "Slow 3G"
   - Player should switch to lower quality

2. **Manual Quality Check**:
   ```bash
   # Check each variant playlist has segments
   ls -la /tmp/hls/{stream_id}/1080p/
   ls -la /tmp/hls/{stream_id}/720p/
   ls -la /tmp/hls/{stream_id}/480p/
   ls -la /tmp/hls/{stream_id}/360p/
   ```

3. **Verify Segment Sync**:
   All variants should have approximately the same number of segments at any given time.

## Performance Considerations

### CPU Usage

ABR generates 4 simultaneous encodes, which significantly increases CPU usage:

- **Without ABR**: ~30-50% CPU (single encode)
- **With ABR**: ~150-200% CPU (4 encodes)

### Hardware Acceleration

To reduce CPU usage, enable hardware acceleration:

```bash
# NVIDIA GPUs
HW_ACCEL_ENABLED=true
HW_ACCEL_TYPE=nvenc

# Intel GPUs
HW_ACCEL_ENABLED=true
HW_ACCEL_TYPE=qsv

# macOS
HW_ACCEL_ENABLED=true
HW_ACCEL_TYPE=videotoolbox
```

## Troubleshooting

### Issue: Master playlist not generated

**Check**: ABR must be enabled in config
```bash
# Verify in logs
[Transcoder] ABR enabled: true
```

### Issue: Some variants missing

**Check**: FFmpeg errors in logs
```bash
# Look for FFmpeg stderr output
[Transcoder] Command: ffmpeg ...
```

**Common causes**:
- Input resolution too low (can't upscale)
- Insufficient CPU/memory
- FFmpeg not installed

### Issue: Player shows single quality only

**Possible causes**:
1. Loading variant playlist directly instead of master.m3u8
2. Player doesn't support ABR (use Video.js or similar)
3. Cache issues - clear browser cache

### Issue: Segments not syncing

**Fix**: Set keyframe interval in OBS to match HLS segment time
```
OBS Settings → Output → Keyframe Interval: 2 seconds
HLS_SEGMENT_TIME=6 (default)
```

## API Endpoints

### Get Playback URL

```bash
curl http://localhost:8083/streams/{stream_id}/playback
```

Response:
```json
{
  "playback_url": "http://localhost:8083/hls/{stream_id}/master.m3u8"
}
```

### Get Stream Status

```bash
curl http://localhost:8083/streams/{stream_id}
```

## File Cleanup

HLS segments are automatically deleted by FFmpeg using the `delete_segments` flag. Only the last `HLS_PLAYLIST_LENGTH` (default: 5) segments are kept for each variant.

To manually clean up after stream ends:

```bash
rm -rf /tmp/hls/{stream_id}
```

## Next Steps

1. **VOD Creation**: Convert live stream to on-demand video
2. **CDN Integration**: Upload segments to MinIO/S3 for scalability
3. **Dynamic Variant Selection**: Adjust variants based on input resolution
4. **Quality Analytics**: Track which qualities viewers use most
5. **Thumbnail Generation**: Create preview thumbnails for each variant

## References

- [HLS Specification](https://tools.ietf.org/html/rfc8216)
- [FFmpeg HLS Documentation](https://ffmpeg.org/ffmpeg-formats.html#hls-2)
- [Video.js Quality Levels](https://github.com/videojs/http-streaming)
