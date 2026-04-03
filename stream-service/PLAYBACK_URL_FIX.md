# Playback URL Issue - Root Cause and Fix

## The Problem

When you call `/streams/:id/playback`, you're getting:
```json
{
    "playback_url": "rtmp://localhost:1935/live/sk_xxx",
    "protocol": "HLS",
    "note": "Use this URL in HLS video player"
}
```

**The RTMP URL is wrong!** It should return the HLS URL:
```json
{
    "playback_url": "http://localhost:8083/hls/{stream_id}/master.m3u8",
    "protocol": "HLS",
    "note": "Use this URL in HLS video player"
}
```

## Root Cause

Your **video service or frontend** is calling the **gRPC endpoint** instead of the HTTP REST endpoint. The gRPC `GetStream` method returns the Stream object which only has:
- `stream_key` (like `sk_9eb675c255f14fa38e60a89ca5da1420`)
- But NO `playback_url` field

Your video service then incorrectly constructs: `rtmp://localhost:1935/live/{stream_key}` from the stream_key.

## The Fix

You have **3 options**:

### Option 1: Quick Fix - Use HTTP API directly (Recommended)

Update your video service to call the HTTP REST endpoint instead of gRPC:

```bash
# Instead of calling gRPC GetStream(), call:
curl http://localhost:8083/streams/{stream_id}/playback
```

Response:
```json
{
  "playback_url": "http://localhost:8083/hls/{stream_id}/master.m3u8",
  "protocol": "HLS",
  "note": "Use this URL in HLS video player"
}
```

### Option 2: Add playback_url to gRPC (Proper Fix)

1. Run the fix script (requires sudo for proto directory):
   ```bash
   ./fix_playback_url.sh
   ```

2. This will:
   - Copy the updated `proto/stream.proto` (with playback_url field)
   - Regenerate protobuf files
   - Show you how to update server.go

3. Manual step - Update `internal/grpc/server.go` line 184-207:
   ```go
   func streamModelToProto(stream *model.Stream) *streampb.Stream {
       protoStream := &streampb.Stream{
           // ... existing fields ...
       }

       // Add these lines before return:
       if stream.IsLive() {
           protoStream.PlaybackUrl = fmt.Sprintf("http://localhost:8083/hls/%s/master.m3u8", stream.ID.String())
       }
       protoStream.RtmpUrl = fmt.Sprintf("rtmp://localhost:1935/live/%s", stream.StreamKey)

       return protoStream
   }
   ```

4. Restart the stream service

### Option 3: Construct HLS URL in your Video Service

In your video service, when you get a stream from gRPC, construct the HLS URL yourself:

```go
// Go example
stream := // ... from gRPC GetStream() ...
playbackURL := fmt.Sprintf("http://localhost:8083/hls/%s/master.m3u8", stream.Id)
```

```typescript
// TypeScript example
const stream = await streamClient.getStream({ streamId });
const playbackURL = `http://localhost:8083/hls/${stream.id}/master.m3u8`;
```

## URL Reference

| Purpose | Correct URL Format | Protocol |
|---------|-------------------|----------|
| **Streamer** (OBS publish) | `rtmp://localhost:1935/live/{stream_key}` | RTMP |
| **Viewer** (playback) | `http://localhost:8083/hls/{stream_id}/master.m3u8` | HLS |

**Key Difference:**
- RTMP uses `stream_key` (secret, for publishing)
- HLS uses `stream_id` (public UUID, for viewing)

## Testing After Fix

1. Start streaming from OBS to RTMP
2. Call the API:
   ```bash
   curl http://localhost:8083/streams/{stream_id}/playback
   ```
3. You should get the HLS URL
4. Open in VLC:
   ```bash
   vlc "http://localhost:8083/hls/{stream_id}/master.m3u8"
   ```

## Quick Test Commands

```bash
# Find your stream ID
psql -U postgres -d stream_service -c "SELECT id, stream_key, status FROM streams LIMIT 5;"

# Get playback URL
STREAM_ID="your-stream-id-here"
curl "http://localhost:8083/streams/$STREAM_ID/playback"

# Play in VLC (if stream is live)
vlc "http://localhost:8083/hls/$STREAM_ID/master.m3u8"
```

## Files Changed

- `proto/stream.proto` - Added playback_url and rtmp_url fields
- `internal/grpc/server.go` - Updated to populate URLs
- `internal/grpc/streampb/*.pb.go` - Regenerated from proto

## Summary

The HTTP REST API at `/streams/:id/playback` works correctly and returns HLS URLs. The issue is that your video service is using the gRPC endpoint which doesn't have a playback_url field. Either switch to using the HTTP endpoint, or add the field to the gRPC response.
