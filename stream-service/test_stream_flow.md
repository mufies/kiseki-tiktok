# How to Test Your Stream

## Step 1: Check if service is running
```bash
curl http://localhost:8083/health
```

## Step 2: Create a stream in the database or via API
Your stream key is: `sk_9eb675c255f14fa38e60a89ca5da1420`

You need to find the corresponding **stream_id** (UUID) from your database:
```sql
SELECT id, stream_key, title, status FROM streams WHERE stream_key = 'sk_9eb675c255f14fa38e60a89ca5da1420';
```

## Step 3: Start streaming with OBS
- **Server**: rtmp://localhost:1935/live
- **Stream Key**: sk_9eb675c255f14fa38e60a89ca5da1420

## Step 4: Play in VLC (use HLS, NOT RTMP)
Replace {stream_id} with the UUID from step 2:
- **URL**: http://localhost:8083/hls/{stream_id}/master.m3u8

Example:
```
http://localhost:8083/hls/550e8400-e29b-41d4-a716-446655440000/master.m3u8
```

## Why VLC won't play RTMP:
- Your RTMP server is "publish-only" (for OBS to send data)
- Viewers must use HLS (HTTP Live Streaming)
- RTMP → Server → FFmpeg → HLS → Viewers
