# 📺 Streaming URLs Quick Reference

## Two Different URLs for Two Different Purposes

### 📤 RTMP URL - For Publishing (OBS → Server)

**Format:**
```
rtmp://localhost:1935/live/{stream_key}
```

**Example:**
```
rtmp://localhost:1935/live/sk_9eb675c255f14fa38e60a89ca5da1420
```

**Used By:**
- OBS Studio
- Streaming software
- FFmpeg (as publisher)
- Mobile streaming apps

**Purpose:**
- Sending video FROM your device TO the server
- OBS uses this to publish your stream
- Server validates the stream_key

**Where to find it:**
- Frontend: After creating stream, shown in OBS setup instructions
- Backend: `stream.stream_key` from database
- Format: RTMP server base URL + stream_key

**⚠️ Important:**
- This is for PUBLISHING only
- Browsers CANNOT play RTMP URLs
- Each stream has a unique stream_key
- Never share your stream_key publicly (like a password)

---

### 📺 HLS URL - For Playback (Server → Viewers)

**Format:**
```
http://localhost:8083/hls/{stream_id}/master.m3u8
```

**Example:**
```
http://localhost:8083/hls/90b20aa1-f4a4-4b92-97bb-9c86cd0da8fb/master.m3u8
```

**Used By:**
- Web browsers (Chrome, Firefox, Safari)
- HLS.js library
- Video players (VLC, ffplay)
- Mobile apps (native HLS support)

**Purpose:**
- Delivering video FROM server TO viewers
- Browsers play this URL
- Supports adaptive bitrate streaming

**Where to find it:**
- API Endpoint: `GET /streams/{stream_id}/playback`
- Direct construction: `http://localhost:8083/hls/{stream_id}/master.m3u8`
- Frontend: Returned from `streamAPI.getPlaybackUrl()`

**⚠️ Important:**
- This is for PLAYBACK only
- OBS CANNOT use HLS URLs for publishing
- Each stream has a unique stream_id (different from stream_key)
- This URL can be shared publicly for viewers

---

## 🔄 Complete Flow

```
┌─────────────┐
│   Streamer  │
│  (You/OBS)  │
└──────┬──────┘
       │
       │ 📤 Publishes to RTMP URL
       │ rtmp://localhost:1935/live/{stream_key}
       │
       ↓
┌─────────────┐
│   Server    │
│  (Backend)  │
└──────┬──────┘
       │
       │ 🎞️ Transcodes to HLS
       │
       ↓
       │
       │ 📺 Serves HLS URL
       │ http://localhost:8083/hls/{stream_id}/master.m3u8
       │
       ↓
┌─────────────┐
│   Viewers   │
│  (Browser)  │
└─────────────┘
```

---

## 📝 How to Get Each URL

### RTMP URL (Publishing)

**Method 1: From Frontend UI**
```javascript
// After creating stream
const stream = await streamAPI.createStream({ user_id, title });
const rtmpUrl = `${RTMP_SERVER}/${stream.stream_key}`;
// rtmp://localhost:1935/live/sk_xxx
```

**Method 2: From Backend API**
```bash
# Create stream
curl -X POST http://localhost:8080/streams \
  -H "Authorization: Bearer TOKEN" \
  -d '{"user_id":"xxx","title":"My Stream"}'

# Response includes stream_key
{
  "stream": {
    "id": "stream-id",
    "stream_key": "sk_xxx"  # ← Use this with RTMP base URL
  }
}

# RTMP URL = rtmp://localhost:1935/live/sk_xxx
```

---

### HLS URL (Playback)

**Method 1: Call API Endpoint (Recommended)**
```bash
curl http://localhost:8080/streams/{stream_id}/playback \
  -H "Authorization: Bearer TOKEN"
```

**Response:**
```json
{
  "playback_url": "http://localhost:8083/hls/{stream_id}/master.m3u8",
  "protocol": "HLS",
  "note": "Use this URL in HLS video player"
}
```

**Method 2: Construct Manually**
```javascript
const hlsUrl = `http://localhost:8083/hls/${stream.id}/master.m3u8`;
```

**Method 3: Frontend Helper**
```javascript
const playback = await streamAPI.getPlaybackUrl(streamId);
console.log(playback.hls_url);
// http://localhost:8083/hls/{stream_id}/master.m3u8
```

---

## 🎯 Key Differences

| Aspect | RTMP URL | HLS URL |
|--------|----------|---------|
| **Purpose** | Publishing (upload) | Playback (download) |
| **Protocol** | RTMP | HTTP/HTTPS |
| **Port** | 1935 | 8083 |
| **Identifier** | stream_key | stream_id |
| **Format** | rtmp://.../{key} | http://.../master.m3u8 |
| **Used By** | OBS, Encoders | Browsers, Players |
| **Direction** | You → Server | Server → Viewers |
| **Security** | Private (secret key) | Public (shareable) |
| **Example** | rtmp://localhost:1935/live/sk_abc123 | http://localhost:8083/hls/uuid-123/master.m3u8 |

---

## ❌ Common Mistakes

### ❌ Using RTMP URL in Browser
```javascript
// WRONG - Browser cannot play RTMP
<video src="rtmp://localhost:1935/live/sk_xxx" />
// Error: ERR_UNKNOWN_URL_SCHEME
```

### ✅ Correct - Use HLS URL in Browser
```javascript
// CORRECT - Browser can play HLS
<video src="http://localhost:8083/hls/stream-id/master.m3u8" />
```

---

### ❌ Using HLS URL in OBS
```
WRONG - OBS cannot publish to HLS endpoint
Server: http://localhost:8083/hls/stream-id/master.m3u8
```

### ✅ Correct - Use RTMP URL in OBS
```
CORRECT - OBS publishes to RTMP
Server: rtmp://localhost:1935/live
Stream Key: sk_xxx
```

---

## 🔧 Environment Variables

```bash
# .env file
VITE_API_URL=http://localhost:8080              # API Gateway
VITE_STREAM_SERVICE_URL=http://localhost:8083   # HLS files location
VITE_RTMP_URL=rtmp://localhost:1935/live        # RTMP server
```

---

## 🧪 Testing

### Test RTMP Publishing
```bash
ffmpeg -re -i video.mp4 \
  -c:v libx264 -c:a aac \
  -f flv rtmp://localhost:1935/live/YOUR_STREAM_KEY
```

### Test HLS Playback
```bash
# Open in VLC or ffplay
ffplay http://localhost:8083/hls/YOUR_STREAM_ID/master.m3u8

# Or curl to check if file exists
curl -I http://localhost:8083/hls/YOUR_STREAM_ID/master.m3u8
```

---

## 📚 Summary

**Remember:**
- **RTMP = Publishing** (OBS → Server) - Uses stream_key
- **HLS = Playback** (Server → Browser) - Uses stream_id
- **RTMP URL** is for streaming software
- **HLS URL** is for video players
- **Never confuse the two!**

**Frontend API:**
```typescript
// Creating stream gives you stream_key for RTMP
const stream = await streamAPI.createStream({...});
const rtmpUrl = `rtmp://localhost:1935/live/${stream.stream_key}`;

// Getting playback gives you HLS URL
const playback = await streamAPI.getPlaybackUrl(stream.id);
const hlsUrl = playback.hls_url; // http://...master.m3u8
```
