# 🚨 HLS File Not Found (404) - Troubleshooting Guide

## Issue Summary

**Symptoms:**
- Frontend correctly gets HLS URL: `http://localhost:8083/hls/{stream_id}/master.m3u8`
- Video player shows: "Failed to load because no supported source was found"
- Console shows: `NotSupportedError`
- Testing the URL returns: `HTTP 404 Not Found`

**Root Cause:**
The HLS transcoder is not generating the `.m3u8` playlist files and `.ts` segment files.

---

## ✅ What's Working

1. ✅ Stream creation (stream_id and stream_key generated)
2. ✅ RTMP URL format is correct
3. ✅ Backend API returns correct HLS URL format
4. ✅ Frontend validates and uses HLS URL (not RTMP)
5. ✅ Stream status changes to "live"

## ❌ What's Not Working

1. ❌ HLS files are not being generated in `/tmp/hls/{stream_id}/`
2. ❌ Accessing HLS URL returns 404
3. ❌ FFmpeg transcoding process might not be running

---

## 🔍 Diagnostic Steps

### Step 1: Check if Stream Service is Running

```bash
# Check if stream service is listening on port 8083
curl http://localhost:8083/health

# Or check with netstat
netstat -tuln | grep 8083

# Expected: Should return health check response or show port is listening
```

### Step 2: Check if RTMP Server is Running

```bash
# Check if RTMP port is listening
netstat -tuln | grep 1935

# Expected: Port 1935 should be listening
```

### Step 3: Test RTMP Connection from OBS

```
OBS Settings:
- Server: rtmp://localhost:1935/live
- Stream Key: sk_xxx (from stream creation)

Start Streaming and check:
- Does OBS show "Live" indicator?
- Does OBS show upload bitrate?
```

### Step 4: Check Stream Service Logs

```bash
# If running via Docker
docker logs stream-service -f

# If running locally
# Check the terminal where you started the stream service
```

**Look for these log messages:**

✅ **Good logs (HLS working):**
```
OnPublish received for stream key: sk_xxx
Validated stream key: sk_xxx
Starting HLS transcoder for stream: stream-id
FFmpeg process started with PID: xxxx
HLS segment written: /tmp/hls/stream-id/segment_0.ts
```

❌ **Bad logs (HLS not working):**
```
OnPublish received but no transcoder started
FFmpeg failed to start
Permission denied writing to /tmp/hls
No such file or directory: /tmp/hls
```

### Step 5: Check HLS Files on Filesystem

```bash
# Check if HLS directory exists
ls -la /tmp/hls/

# Check specific stream directory
ls -la /tmp/hls/YOUR_STREAM_ID/

# Expected files:
# - master.m3u8 (master playlist)
# - variant playlists (1080p.m3u8, 720p.m3u8, etc.)
# - .ts segment files (segment_0.ts, segment_1.ts, etc.)
```

### Step 6: Test HLS URL Directly

```bash
# Test with curl
curl -I http://localhost:8083/hls/YOUR_STREAM_ID/master.m3u8

# Test with ffplay (if installed)
ffplay http://localhost:8083/hls/YOUR_STREAM_ID/master.m3u8

# Test in browser (open URL directly)
```

---

## 🛠️ Common Causes & Solutions

### Cause 1: HLS Transcoder Not Starting

**Symptoms:**
- Stream status is "live" but no HLS files generated
- No FFmpeg process in logs

**Solutions:**

1. **Check FFmpeg Installation**
   ```bash
   which ffmpeg
   ffmpeg -version
   ```
   If not installed: `sudo apt install ffmpeg` (Ubuntu/Debian)

2. **Check Stream Service Code**
   - Look at `internal/rtmp/handler.go` - OnPublish function
   - Verify it calls the HLS transcoder
   - Check `internal/transcoder/manager.go` - Start function

3. **Check Transcoder Configuration**
   - File: `config/config.go` or environment variables
   - Verify HLS output directory is set correctly
   - Default should be: `/tmp/hls`

### Cause 2: Permission Issues

**Symptoms:**
- Logs show "Permission denied"
- Cannot create `/tmp/hls` directory

**Solutions:**

```bash
# Create directory with correct permissions
sudo mkdir -p /tmp/hls
sudo chmod 777 /tmp/hls

# Or run stream service with appropriate user
```

### Cause 3: Wrong Stream Status

**Symptoms:**
- Stream status is "offline" or "ending"
- HLS files were deleted

**Solutions:**

1. **Check stream status:**
   ```bash
   curl http://localhost:8080/streams/YOUR_STREAM_ID \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

2. **Make sure OBS is still streaming**
   - Check OBS shows "Live"
   - Check bitrate is uploading

3. **Restart OBS stream if needed**

### Cause 4: Port 8083 Not Accessible

**Symptoms:**
- Cannot reach http://localhost:8083
- Connection refused

**Solutions:**

1. **Check if stream service is running:**
   ```bash
   ps aux | grep stream-service
   ```

2. **Check firewall:**
   ```bash
   sudo ufw status
   sudo ufw allow 8083
   ```

3. **Start stream service if not running**

### Cause 5: CORS Issues

**Symptoms:**
- Browser console shows CORS error
- Network tab shows preflight OPTIONS failed

**Solutions:**

Check stream service CORS configuration:
```go
// Should have these headers
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

### Cause 6: FFmpeg Command Issues

**Symptoms:**
- FFmpeg starts but immediately crashes
- Logs show FFmpeg error

**Solutions:**

1. **Test FFmpeg command manually:**
   ```bash
   # Try running the transcoding command manually
   ffmpeg -i pipe:0 \
     -c:v libx264 -preset veryfast -g 50 \
     -c:a aac -b:a 128k \
     -f hls -hls_time 6 -hls_list_size 5 \
     -hls_flags delete_segments \
     /tmp/hls/test/master.m3u8
   ```

2. **Check FFmpeg supports required codecs:**
   ```bash
   ffmpeg -codecs | grep h264
   ffmpeg -codecs | grep aac
   ```

3. **Update FFmpeg if needed:**
   ```bash
   sudo apt update
   sudo apt install ffmpeg
   ```

---

## 📋 Checklist for HLS to Work

- [ ] Stream service is running
- [ ] Port 8083 is accessible (HTTP server)
- [ ] Port 1935 is accessible (RTMP server)
- [ ] FFmpeg is installed and accessible
- [ ] `/tmp/hls` directory exists with write permissions
- [ ] OBS is connected to RTMP server
- [ ] OBS is actively streaming (shows bitrate)
- [ ] Stream status in database is "live"
- [ ] HLS transcoder is started on RTMP connection
- [ ] FFmpeg process is running
- [ ] HLS files are being written to `/tmp/hls/{stream_id}/`
- [ ] HTTP server serves files from `/tmp/hls/`

---

## 🔧 Quick Fixes

### Fix 1: Restart Everything

```bash
# Stop OBS streaming
# Stop stream service
# Clear HLS directory
sudo rm -rf /tmp/hls/*

# Start stream service
# Start OBS streaming
# Wait 10 seconds
# Test HLS URL
```

### Fix 2: Manual HLS Directory Setup

```bash
# Create HLS directory
sudo mkdir -p /tmp/hls
sudo chmod 777 /tmp/hls

# Restart stream service
```

### Fix 3: Test with Simple Stream

```bash
# Generate test video
ffmpeg -f lavfi -i testsrc=duration=60:size=1280x720:rate=30 \
  -f lavfi -i sine=frequency=1000:duration=60 \
  -pix_fmt yuv420p test.mp4

# Stream to RTMP
ffmpeg -re -i test.mp4 \
  -c:v libx264 -preset veryfast \
  -c:a aac \
  -f flv rtmp://localhost:1935/live/YOUR_STREAM_KEY

# Wait 10 seconds
# Test HLS URL
curl http://localhost:8083/hls/YOUR_STREAM_ID/master.m3u8
```

---

## 🎯 Expected Working Flow

1. **OBS connects to RTMP:**
   ```
   OBS → rtmp://localhost:1935/live/{stream_key}
   ```

2. **Stream service receives RTMP:**
   ```
   RTMP Handler validates stream_key
   Updates stream status to "live"
   Starts HLS transcoder
   ```

3. **FFmpeg transcodes to HLS:**
   ```
   FFmpeg receives RTMP stream
   Generates master.m3u8
   Generates variant playlists
   Writes .ts segments to /tmp/hls/{stream_id}/
   ```

4. **HTTP server serves HLS:**
   ```
   Browser requests: http://localhost:8083/hls/{stream_id}/master.m3u8
   Server responds: 200 OK with playlist
   Browser loads segments
   Video plays
   ```

---

## 🆘 Still Not Working?

### Check Stream Service Implementation

The stream service needs these components:

1. **RTMP Handler** (`internal/rtmp/handler.go`)
   - OnPublish function
   - Stream key validation
   - HLS transcoder startup

2. **HLS Transcoder** (`internal/transcoder/hls_transcoder.go`)
   - FFmpeg command construction
   - Process management
   - File output handling

3. **HTTP Server** (`internal/handler/stream_handler.go`)
   - `/hls/*filepath` endpoint
   - Static file serving
   - Correct CORS headers

### Enable Debug Logging

Add verbose logging to see what's happening:

```go
// In RTMP handler OnPublish
log.Printf("DEBUG: OnPublish called for stream key: %s", streamKey)
log.Printf("DEBUG: Starting HLS transcoder for stream: %s", stream.ID)

// In HLS transcoder Start
log.Printf("DEBUG: FFmpeg command: %v", cmd.Args)
log.Printf("DEBUG: HLS output directory: %s", outputDir)
```

### Contact Backend Team

If frontend is working but backend isn't generating HLS files, provide this info:

- Stream ID
- Stream key
- Stream status
- HLS URL that returns 404
- Stream service logs
- FFmpeg installation details
- Directory permissions

---

## ✅ Frontend Improvements (Already Done)

The frontend now:
- ✅ Tests HLS URL before playing
- ✅ Shows clear error messages
- ✅ Provides "Test HLS URL" button in debug panel
- ✅ Logs all API responses
- ✅ Handles 404 errors gracefully
- ✅ Suggests waiting for transcoding

The issue is **100% backend** - the HLS transcoding is not working.
