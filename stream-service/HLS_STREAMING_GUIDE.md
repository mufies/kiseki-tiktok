# HLS Streaming Guide

## Tổng quan

Service này đã được tích hợp FFmpeg transcoder để convert RTMP stream từ OBS sang HLS (HTTP Live Streaming) cho frontend.

## Kiến trúc

```
OBS → RTMP Server → FLV Muxer → FFmpeg → HLS Segments → Frontend Player
```

### Flow hoạt động:

1. **OBS stream vào RTMP server** (port 1935)
2. **RTMP handler** nhận audio/video packets
3. **FLV Writer** convert packets sang FLV format
4. **FFmpeg** đọc FLV từ stdin, transcode thành HLS
5. **HLS segments** (.ts) và playlist (.m3u8) được lưu tại `/tmp/hls/{stream_id}/`
6. **Frontend** fetch HLS playlist qua HTTP endpoint

## Cấu hình

### Environment Variables

```bash
# Transcoding
ENABLE_TRANSCODING=true
FFMPEG_PATH=ffmpeg  # hoặc đường dẫn đầy đủ: /usr/bin/ffmpeg

# HLS Settings
HLS_SEGMENT_TIME=6        # Độ dài mỗi segment (giây)
HLS_PLAYLIST_LENGTH=5     # Số segments giữ trong playlist
```

### FFmpeg Settings (trong code)

```go
// internal/transcoder/hls_transcoder.go
-c:v copy           // Copy video codec (không re-encode, low latency)
-c:a aac            // Audio codec AAC
-b:a 128k           // Audio bitrate
-f hls              // HLS output
-hls_time 6         // 6 giây/segment
-hls_list_size 5    // Giữ 5 segments
-hls_flags delete_segments  // Tự động xóa segments cũ
```

## Sử dụng

### 1. Start Service

```bash
# Đảm bảo FFmpeg đã cài
ffmpeg -version

# Chạy service
go run cmd/main.go
```

### 2. Stream từ OBS

#### OBS Settings:
```
Server: rtmp://localhost:1935/live
Stream Key: <your_stream_key_from_database>
```

#### Advanced Settings:
- Keyframe Interval: 2 seconds (khuyến nghị cho HLS)
- Video Bitrate: 2500 kbps
- Audio Bitrate: 128 kbps

### 3. Access HLS Stream

Khi stream started, xem logs:
```
[Transcoder] ✅ FFmpeg started for stream {stream_id}
[RTMP] 📹 HLS transcoder started - Playlist: /hls/{stream_id}/playlist.m3u8
```

#### HTTP Endpoint:
```
http://localhost:8083/hls/{stream_id}/playlist.m3u8
```

### 4. Frontend Integration

#### Sử dụng Video.js:

```html
<video id="video-player" class="video-js" controls>
  <source src="http://localhost:8083/hls/{stream_id}/playlist.m3u8" type="application/x-mpegURL">
</video>

<script src="https://vjs.zencdn.net/7.20.3/video.min.js"></script>
<script>
  var player = videojs('video-player', {
    liveui: true,
    controls: true,
    autoplay: true,
    preload: 'auto'
  });
</script>
```

#### Sử dụng HLS.js:

```javascript
import Hls from 'hls.js';

const video = document.getElementById('video');
const hlsUrl = `http://localhost:8083/hls/${streamId}/playlist.m3u8`;

if (Hls.isSupported()) {
  const hls = new Hls({
    enableWorker: true,
    lowLatencyMode: true,
    backBufferLength: 90
  });

  hls.loadSource(hlsUrl);
  hls.attachMedia(video);

  hls.on(Hls.Events.MANIFEST_PARSED, () => {
    video.play();
  });
}
```

#### Native HLS (Safari):

```javascript
// Safari hỗ trợ HLS native
if (video.canPlayType('application/vnd.apple.mpegurl')) {
  video.src = hlsUrl;
  video.play();
}
```

## File Structure

```
/tmp/hls/
├── {stream_id_1}/
│   ├── playlist.m3u8      # Master playlist
│   ├── segment_000.ts     # Video segment 1
│   ├── segment_001.ts     # Video segment 2
│   └── segment_002.ts     # Video segment 3
├── {stream_id_2}/
│   ├── playlist.m3u8
│   └── segment_*.ts
```

## API Endpoints

### Get HLS Playlist URL

```bash
GET /streams/:id/playback

Response:
{
  "stream_id": "uuid",
  "status": "live",
  "hls_url": "/hls/{stream_id}/playlist.m3u8",
  "viewers": 42
}
```

### Get Stream Info

```bash
GET /streams/:id

Response:
{
  "id": "uuid",
  "title": "My Stream",
  "status": "live",
  "hls_url": "/hls/{stream_id}/playlist.m3u8"
}
```

## Troubleshooting

### FFmpeg không start

**Lỗi:** `Failed to start transcoder: exec: "ffmpeg": executable file not found`

**Fix:**
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# macOS
brew install ffmpeg

# Hoặc set đường dẫn cụ thể
export FFMPEG_PATH=/usr/local/bin/ffmpeg
```

### HLS segments không tạo ra

**Check:**
1. FFmpeg có quyền write vào `/tmp/hls/`
```bash
ls -la /tmp/hls/
```

2. Check FFmpeg logs (stdout/stderr)
```bash
# Logs sẽ show trong console khi stream
```

### Latency cao

**Giải pháp:**
1. Giảm `HLS_SEGMENT_TIME` (từ 6s → 2s)
2. Sử dụng Low-Latency HLS (LL-HLS) - requires FFmpeg 4.3+
3. Tune OBS settings:
   - Giảm Keyframe Interval
   - Enable CBR (Constant Bitrate)

### Video lag/buffer

**Check:**
1. Network bandwidth
2. FFmpeg CPU usage
3. Giảm video bitrate trong OBS
4. Enable hardware encoding (NVENC/QuickSync)

## Performance Tips

### 1. Hardware Acceleration

Sửa FFmpeg args trong `hls_transcoder.go`:

```go
// NVIDIA GPU
"-c:v", "h264_nvenc",

// Intel QuickSync
"-c:v", "h264_qsv",

// Apple VideoToolbox (macOS)
"-c:v", "h264_videotoolbox",
```

### 2. Multiple Quality Levels (Adaptive Bitrate)

Tạo multiple transcoders với different bitrates:

```go
// 1080p, 720p, 480p, 360p
// TODO: Implement multi-bitrate transcoding
```

### 3. Cleanup Old Segments

Hiện tại FFmpeg tự động xóa với flag `-hls_flags delete_segments`.

Nếu muốn manual cleanup:
```bash
# Cronjob xóa HLS files cũ hơn 1 giờ
0 * * * * find /tmp/hls -type d -mmin +60 -exec rm -rf {} \;
```

## Monitoring

### Check Active Transcoders

```bash
# Endpoint (TODO: implement)
GET /admin/transcoders

Response:
{
  "active_count": 3,
  "transcoders": [
    {
      "stream_id": "uuid",
      "is_running": true,
      "duration": 123.45,
      "bitrate_mbps": 2.5,
      "playlist_url": "/hls/uuid/playlist.m3u8"
    }
  ]
}
```

### System Monitoring

```bash
# Check FFmpeg processes
ps aux | grep ffmpeg

# Monitor CPU/Memory
top -p $(pgrep -d',' ffmpeg)

# Check HLS directory size
du -sh /tmp/hls/*
```

## Roadmap

- [ ] Multi-bitrate ABR (Adaptive Bitrate)
- [ ] Low-Latency HLS (LL-HLS)
- [ ] DVR/Timeshift support
- [ ] Upload HLS to MinIO/S3 for CDN
- [ ] Thumbnail generation
- [ ] VOD conversion after stream ends
- [ ] Quality analytics
- [ ] GPU acceleration auto-detection

## Security Notes

### Production Recommendations:

1. **CORS**: Cấu hình CORS cho HLS endpoints
2. **Authentication**: Protect HLS URLs with tokens
3. **Rate Limiting**: Prevent HLS endpoint abuse
4. **CDN**: Serve HLS through CDN (CloudFront, CloudFlare)
5. **DRM**: Implement AES-128 encryption for premium content

### Token-based HLS URLs

```go
// TODO: Implement signed URLs
/hls/{stream_id}/playlist.m3u8?token={jwt_token}
```

## Testing

### Test HLS Playback

```bash
# Test với ffplay
ffplay http://localhost:8083/hls/{stream_id}/playlist.m3u8

# Test với VLC
vlc http://localhost:8083/hls/{stream_id}/playlist.m3u8

# Test với curl
curl http://localhost:8083/hls/{stream_id}/playlist.m3u8
```

### Load Testing

```bash
# Simulate multiple viewers
for i in {1..10}; do
  ffplay -nodisp -autoexit http://localhost:8083/hls/{stream_id}/playlist.m3u8 &
done
```

## Support

Có vấn đề? Check logs:
```bash
# RTMP logs
[RTMP] prefix messages

# Transcoder logs
[Transcoder] prefix messages

# FFmpeg stderr
FFmpeg output in console
```
