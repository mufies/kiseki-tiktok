# Stream Service

A microservice for live streaming with RTMP ingestion and HLS transcoding.

## Features

- RTMP server for stream ingestion (OBS, etc.)
- Automatic HLS transcoding using FFmpeg
- Multi-stream support with isolated statistics
- Real-time viewer tracking
- gRPC and HTTP APIs
- MinIO integration for VOD storage
- Kafka event publishing
- Redis for real-time data

## Architecture

```
OBS/Streaming Software (RTMP)
    -> RTMP Server (port 1935)
    -> FFmpeg Transcoder
    -> HLS Output (/tmp/hls/{stream_id}/)
    -> HTTP Server (port 8083)
    -> Frontend Player
```

## Prerequisites

- Go 1.21+
- FFmpeg
- PostgreSQL
- Redis (optional)
- Kafka (optional)
- MinIO (optional)

## Installation

### Install FFmpeg

```bash
# Ubuntu/Debian
sudo apt install ffmpeg

# macOS
brew install ffmpeg

# Verify
ffmpeg -version
```

### Install Dependencies

```bash
go mod download
```

## Configuration

Create a `.env` file or set environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=stream_service

# RTMP
RTMP_PORT=1935

# HLS
HLS_SEGMENT_TIME=6
HLS_PLAYLIST_LENGTH=5

# Transcoding
ENABLE_TRANSCODING=true
FFMPEG_PATH=ffmpeg

# HTTP Server
SERVER_PORT=8083
GRPC_PORT=50055

# Redis (optional)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka (optional)
KAFKA_BROKERS=localhost:9092

# MinIO (optional)
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_STREAMS_BUCKET=streams
```

## Running

### Development

```bash
go run cmd/main.go
```

### Production

```bash
go build -o bin/stream-service cmd/main.go
./bin/stream-service
```

## Usage

### 1. Create a Stream

```bash
curl -X POST http://localhost:8083/streams \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "title": "My Live Stream",
    "description": "Stream description"
  }'
```

Response includes `stream_id` and `stream_key`.

### 2. Stream from OBS

Configure OBS Studio:
- Server: `rtmp://localhost:1935/live`
- Stream Key: `{stream_key}` from step 1

### 3. Watch Stream

HLS URL:
```
http://localhost:8083/hls/{stream_id}/playlist.m3u8
```

Use any HLS-compatible player (Video.js, HLS.js, VLC, etc.)

## API Endpoints

### Streams

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /streams | Create stream |
| GET | /streams/:id | Get stream details |
| PATCH | /streams/:id | Update stream |
| DELETE | /streams/:id | Delete stream |
| POST | /streams/:id/start | Start stream |
| POST | /streams/:id/end | End stream |
| GET | /streams/live | List live streams |
| GET | /streams/:id/playback | Get playback URL |
| POST | /streams/:id/viewers/join | Join as viewer |
| POST | /streams/:id/viewers/leave | Leave stream |

### HLS Files

- GET /hls/{stream_id}/playlist.m3u8 - Master playlist
- GET /hls/{stream_id}/segment_*.ts - Video segments

### Health Check

- GET /health - Service health status

## gRPC API

Port: 50055

Services:
- GetStream(stream_id)
- GetStreamByKey(stream_key)
- ListLiveStreams()

## Testing

### Run Tests

```bash
go test ./...
```

### Test HLS Streaming

```bash
./test_hls.sh
```

### Test with HTML Player

```bash
open test_player.html
```

### Test with ffplay

```bash
ffplay http://localhost:8083/hls/{stream_id}/playlist.m3u8
```

## Database Schema

```sql
CREATE TABLE streams (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    stream_key VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    viewer_count INT DEFAULT 0,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Project Structure

```
stream-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handler/                # HTTP handlers
│   ├── service/                # Business logic
│   ├── repository/             # Database layer
│   ├── model/                  # Data models
│   ├── rtmp/                   # RTMP server
│   │   ├── server.go
│   │   └── handler.go
│   ├── transcoder/             # HLS transcoding
│   │   ├── hls_transcoder.go
│   │   ├── manager.go
│   │   └── flv_writer.go
│   ├── grpc/                   # gRPC server
│   ├── kafka/                  # Kafka producer
│   └── storage/                # MinIO client
├── config/                     # Configuration
├── proto/                      # Protobuf definitions
├── examples/                   # Example code
├── bin/                        # Compiled binaries
└── test_player.html           # Test player
```

## Performance

### Resource Usage (per stream)

- CPU: 5-15% (with video copy)
- Memory: 50-100 MB
- Disk: ~10 MB (temporary HLS files)

### Scalability

- Tested: 10 concurrent streams
- Recommended: 3-5 streams per CPU core
- Latency: 7-10 seconds (tunable)

## Troubleshooting

### FFmpeg not found

```bash
export FFMPEG_PATH=/usr/local/bin/ffmpeg
```

### No HLS segments

- Check FFmpeg logs in console output
- Wait 6-12 seconds after stream starts
- Verify /tmp/hls/ directory exists and is writable

### High CPU usage

Enable hardware encoding in `internal/transcoder/hls_transcoder.go`:

```go
// NVIDIA GPU
"-c:v", "h264_nvenc"

// Intel QuickSync
"-c:v", "h264_qsv"

// Apple VideoToolbox
"-c:v", "h264_videotoolbox"
```

## Development

### Run with live reload

```bash
go install github.com/cosmtrek/air@latest
air
```

### Database migrations

Migrations are auto-run on startup using GORM AutoMigrate.

### Generate protobuf

```bash
protoc --go_out=. --go-grpc_out=. proto/stream.proto
```

## Production Deployment

### Docker

```bash
docker build -t stream-service .
docker run -p 1935:1935 -p 8083:8083 stream-service
```

### Docker Compose

```bash
docker-compose up -d
```

### Security Considerations

- Add authentication to HLS URLs
- Implement rate limiting
- Restrict CORS to specific domains
- Use HTTPS/TLS in production
- Validate user permissions
- Monitor for abuse

## Roadmap

### Phase 1 (Complete)
- RTMP server
- HLS transcoding
- Multi-stream support
- HTTP API

### Phase 2 (In Progress)
- Multi-bitrate ABR streaming
- GPU acceleration
- VOD conversion
- CDN integration

### Phase 3 (Planned)
- Low-latency HLS
- WebRTC support
- DVR/Timeshift
- Analytics dashboard

## Documentation

- [HLS Streaming Guide](HLS_STREAMING_GUIDE.md) - Detailed HLS documentation
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Technical details
- [RTMP Implementation](RTMP_IMPLEMENTATION.md) - RTMP server details

## Contributing

1. Fork the repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## License

MIT License

## Support

For issues and questions:
- Check documentation files
- Run `./test_hls.sh` for diagnostics
- Review logs for errors
- Create GitHub issue with details
