# RTMP Server Implementation

## Tổng quan

RTMP (Real-Time Messaging Protocol) server cho phép accept live streams từ OBS Studio, FFmpeg, hoặc bất kỳ RTMP encoder nào.

## Files

- **handler.go** - RTMP event handler, integrate với StreamService
- **server.go** - RTMP server wrapper

## Flow hoạt động

```
1. OBS/FFmpeg connects → RTMP Server
2. RTMP Handshake
3. OnConnect → Validate app
4. OnCreateStream → Create stream channel
5. OnPublish → Validate stream_key, start stream
6. OnVideo/OnAudio → Receive stream data
7. OnClose → End stream, publish events
```

## Cách test

### 1. Tạo stream qua API

```bash
curl -X POST http://localhost:8083/streams \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "My Live Stream",
    "description": "Testing RTMP"
  }'
```

Response sẽ chứa `stream_key`, ví dụ: `sk_abc123xyz`

### 2. Stream với OBS Studio

**Settings → Stream:**
- Service: **Custom**
- Server: `rtmp://localhost:1935/live`
- Stream Key: `sk_abc123xyz` (từ API response)

**Start Streaming!**

### 3. Stream với FFmpeg

```bash
# Stream từ file
ffmpeg -re -i video.mp4 \
  -c:v libx264 -preset veryfast -b:v 2500k \
  -c:a aac -b:a 128k \
  -f flv rtmp://localhost:1935/live/sk_abc123xyz

# Stream từ webcam (Linux)
ffmpeg -f v4l2 -i /dev/video0 \
  -f alsa -i hw:0 \
  -c:v libx264 -preset ultrafast \
  -c:a aac \
  -f flv rtmp://localhost:1935/live/sk_abc123xyz

# Stream từ webcam (macOS)
ffmpeg -f avfoundation -i "0:0" \
  -c:v libx264 -preset ultrafast \
  -c:a aac \
  -f flv rtmp://localhost:1935/live/sk_abc123xyz
```

### 4. Verify stream status

```bash
# Check stream đang live
curl http://localhost:8083/streams/live

# Check specific stream
curl http://localhost:8083/streams/{stream-id}
```

## StreamHandler Methods

### Connection Lifecycle

| Method | Description |
|--------|-------------|
| `OnServe` | Client kết nối TCP |
| `OnConnect` | RTMP connect command |
| `OnCreateStream` | Tạo stream channel |
| `OnClose` | Connection đóng |

### Publishing Lifecycle

| Method | Description |
|--------|-------------|
| `OnReleaseStream` | Flash compatibility - release stream |
| `OnFCPublish` | Flash compatibility - before publish |
| `OnPublish` | **Start publishing** - validate stream_key |
| `OnSetDataFrame` | Receive metadata (resolution, codec, etc.) |
| `OnVideo` | Receive video packets |
| `OnAudio` | Receive audio packets |
| `OnFCUnpublish` | Flash compatibility - after unpublish |
| `OnDeleteStream` | Delete stream channel |

### Playback (Not Implemented)

| Method | Description |
|--------|-------------|
| `OnPlay` | Client muốn xem stream (returns error) |

### Unknown Messages

| Method | Description |
|--------|-------------|
| `OnUnknownMessage` | Unknown message type |
| `OnUnknownCommandMessage` | Unknown command |
| `OnUnknownDataMessage` | Unknown data |

## ActiveStream Tracking

Handler track active streams trong memory với struct:

```go
type ActiveStream struct {
    StreamID      uuid.UUID
    StreamKey     string
    UserID        uuid.UUID
    StartTime     time.Time
    VideoPackets  int64
    AudioPackets  int64
    TotalBytes    int64
    LastPacketAt  time.Time
}
```

**Statistics được log:**
- Mỗi 100 audio packets
- Mỗi 30 video packets (~1s at 30fps)
- Duration, bitrate, packet counts

## Integration với StreamService

### OnPublish
```go
1. GetStreamByKey(streamKey) → Validate từ database
2. StartStream(streamID) → Update status to "live"
3. Kafka event: stream.started
4. Track trong activeStreams map
```

### OnClose
```go
1. Calculate statistics (duration, bitrate, packets)
2. EndStream(streamID) → Update status to "offline"
3. Kafka event: stream.ended
4. Remove từ activeStreams map
```

## Configuration

Trong `config/config.go`:

```go
RTMPPort:          "1935"
RTMPChunkSize:     4096
RTMPMaxConns:      1000
HLSSegmentTime:    6
HLSPlaylistLength: 5
```

## Logs

Khi stream, bạn sẽ thấy logs:

```
[RTMP] Server listening on :1935
[RTMP] Incoming connection from 192.168.1.100:54321
[RTMP] OnConnect - App: live, FlashVer: FMLE/3.0
[RTMP] OnCreateStream
[RTMP] OnReleaseStream: sk_abc123
[RTMP] OnFCPublish
[RTMP] 📡 OnPublish - Stream Key: sk_abc123, Type: live
[RTMP] ✅ Stream key validated - Stream ID: xxx, User ID: yyy
[RTMP] 🎬 Stream started successfully - ID: xxx, Key: sk_abc123
[RTMP] 📹 Stream sk_abc123 - Video: 30 pkts, Audio: 100 pkts, Duration: 1s, Bitrate: 2.5 Mbps
[RTMP] 🎵 Stream sk_abc123 - Audio: 200 pkts, Video: 60 pkts, Duration: 2s, Bitrate: 2.5 Mbps
...
[RTMP] 🔌 Connection closing
[RTMP] 🛑 Ending stream - ID: xxx, Key: sk_abc123
[RTMP]    Duration: 1m30s
[RTMP]    Video Packets: 2700
[RTMP]    Audio Packets: 9000
[RTMP]    Total Data: 28.5 MB
[RTMP]    Avg Bitrate: 2.53 Mbps
[RTMP] ✅ Stream ended successfully - ID: xxx
```

## TODO - Next Steps

### 1. HLS Transcoding
```
OnVideo/OnAudio → FFmpeg → HLS Segments → MinIO
```

### 2. Recording
```
Save stream to FLV/MP4 → Upload to video-service → VOD
```

### 3. Multi-bitrate
```
Transcode to multiple qualities: 1080p, 720p, 480p, 360p
```

### 4. CDN Integration
```
Push HLS segments to CDN for global distribution
```

## Troubleshooting

### Stream key invalid
```
[RTMP] ❌ Invalid stream key: sk_xxx
```
→ Kiểm tra stream_key có tồn tại trong database không

### Stream cannot go live
```
[RTMP] ❌ Stream cannot go live - Current status: live
```
→ Stream đang live rồi, end stream cũ trước

### Connection drops immediately
```
Check:
- RTMP port 1935 có mở không
- Firewall rules
- Stream key đúng format (sk_xxx)
```

### No video/audio packets
```
Check:
- OBS encoding settings
- Bitrate không quá cao
- Internet connection stable
```

## Performance Notes

**Memory Usage:**
- Each stream: ~minimal overhead (just tracking struct)
- Video/Audio data: Read and discard (not stored in memory)
- TODO: When implementing transcoding, use streaming buffers

**CPU Usage:**
- Current: Very low (just forwarding packets)
- With transcoding: High (FFmpeg processing)

**Network:**
- Bandwidth = stream bitrate × number of publishers
- Typical: 2-5 Mbps per stream

## Security Considerations

**Current:**
- ✅ Stream key authentication
- ✅ Database validation
- ✅ Ownership check

**TODO:**
- Rate limiting (max concurrent streams per user)
- IP whitelisting
- TLS/SSL encryption (RTMPS)
- Token-based authentication
