# Stream Service - Docker Deployment

## ✅ Đã hoàn thành

### 1. gRPC Implementation
- ✅ `proto/stream.proto` - Protocol definition
- ✅ `internal/grpc/streampb/` - Generated gRPC code
- ✅ `internal/grpc/server.go` - gRPC server implementation

**gRPC Methods:**
- GetStream - Get stream by ID
- GetStreamByKey - Get stream by key (for RTMP auth)
- GetLiveStreams - List all live streams
- GetUserStreams - Get user's streams
- UpdateStreamStatus - Update stream status
- GetStreamStats - Get stream statistics

### 2. Docker Configuration
- ✅ `Dockerfile.dev` - Development Dockerfile
- ✅ Updated `docker-compose.yml` - Added stream-service
- ✅ Updated `scripts/init-databases.sh` - Added streamdb

### 3. Service Integration
- ✅ HTTP API on port 8083
- ✅ gRPC server on port 50055
- ✅ RTMP server on port 1935
- ✅ Connected to PostgreSQL, Redis, Kafka, MinIO

## 🚀 How to Run

### Option 1: Run Infrastructure + All Services

```bash
cd /home/mufies/Code/tiktok-clone

# 1. Start infrastructure (PostgreSQL, Redis, Kafka, MinIO)
docker-compose -f docker-compose.infrastructure.yml up -d

# Wait for services to be healthy (~30 seconds)
docker-compose -f docker-compose.infrastructure.yml ps

# 2. Start all application services
docker-compose up -d

# 3. Check stream service logs
docker-compose logs -f stream-service

# 4. Check all services
docker-compose ps
```

### Option 2: Run Only Stream Service (for development)

```bash
# 1. Start infrastructure first
docker-compose -f docker-compose.infrastructure.yml up -d

# 2. Start only stream service
docker-compose up stream-service

# Or run locally:
cd stream-service
go run cmd/main.go
```

## 📊 Service Endpoints

### HTTP API
```
http://localhost:8083/health
http://localhost:8083/streams
```

### gRPC
```
localhost:50055
```

### RTMP (for OBS/FFmpeg)
```
rtmp://localhost:1935/live
```

## 🧪 Testing

### 1. Health Check
```bash
curl http://localhost:8083/health
```

### 2. Create Stream
```bash
curl -X POST http://localhost:8083/streams \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "My Docker Stream",
    "description": "Testing in Docker"
  }'
```

### 3. Stream with OBS
```
Server: rtmp://localhost:1935/live
Key: <stream_key from API>
```

### 4. Check Live Streams
```bash
curl http://localhost:8083/streams/live
```

### 5. Test gRPC (with grpcurl)
```bash
# Install grpcurl first
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50055 list

# Call GetLiveStreams
grpcurl -plaintext \
  -d '{"limit": 10, "offset": 0}' \
  localhost:50055 \
  stream.StreamService/GetLiveStreams
```

## 🔍 Troubleshooting

### Check if service is running
```bash
docker-compose ps stream-service
```

### View logs
```bash
docker-compose logs stream-service
docker-compose logs -f stream-service  # Follow
```

### Restart service
```bash
docker-compose restart stream-service
```

### Rebuild service
```bash
docker-compose build stream-service
docker-compose up -d stream-service
```

### Check database
```bash
docker exec -it tiktok-postgres psql -U postgres -d streamdb -c "SELECT * FROM streams;"
```

### Check Redis
```bash
docker exec -it tiktok-redis redis-cli
# In redis-cli:
KEYS stream:*
GET stream:<id>:viewer_count
```

### Check Kafka
```bash
docker exec -it tiktok-kafka kafka-topics --list --bootstrap-server localhost:9092
```

### Check MinIO
```bash
# Access MinIO console
http://localhost:9010
# Username: minioadmin
# Password: minioadmin
```

## 🌐 Service Dependencies

Stream service connects to:
- **PostgreSQL** (streamdb) - Store stream metadata
- **Redis** - Real-time viewer counts
- **Kafka** - Publish stream events
- **MinIO** - Store HLS segments (future)

External service integration:
- **User Service** (gRPC) - Get user info (future)
- **Video Service** (gRPC) - Create VOD (future)
- **Notification Service** (gRPC) - Notify followers (future)

## 📝 Environment Variables

Configured in docker-compose.yml:

```yaml
# Database
DB_HOST: postgres
DB_PORT: "5432"
DB_USER: postgres
DB_PASSWORD: postgres
DB_NAME: streamdb

# MinIO
MINIO_ENDPOINT: minio:9000
MINIO_ACCESS_KEY: minioadmin
MINIO_SECRET_KEY: minioadmin

# Redis
REDIS_ADDR: redis:6379

# Kafka
KAFKA_BROKERS: kafka:9092

# Ports
SERVER_PORT: "8083"
GRPC_PORT: "50055"
RTMP_PORT: "1935"
```

## 🎯 Port Mapping

| Container Port | Host Port | Purpose |
|---------------|-----------|---------|
| 8083 | 8083 | HTTP API |
| 50055 | 50055 | gRPC |
| 1935 | 1935 | RTMP (live streaming) |

## 🔄 Update Workflow

When you make code changes:

```bash
# Option 1: Auto-reload (if volume mounted)
# Just save file, container will reload

# Option 2: Manual rebuild
docker-compose build stream-service
docker-compose up -d stream-service

# Option 3: Run locally for faster iteration
cd stream-service
go run cmd/main.go
```

## 🛑 Stop Services

```bash
# Stop all services
docker-compose down

# Stop infrastructure
docker-compose -f docker-compose.infrastructure.yml down

# Remove volumes (clean slate)
docker-compose down -v
docker-compose -f docker-compose.infrastructure.yml down -v
```

## 📦 What's Included

✅ HTTP REST API (11 endpoints)
✅ gRPC Server (6 methods)
✅ RTMP Server (live streaming ingest)
✅ PostgreSQL integration
✅ Redis integration
✅ Kafka integration
✅ MinIO integration
✅ Graceful shutdown
✅ Health checks
✅ Logging

## 🔜 Next Steps

After deployment:
1. Implement HLS transcoding
2. Add VOD generation
3. Integrate with user-service (get user info for streams)
4. Integrate with notification-service (notify followers)
5. Add thumbnail generation
6. Implement stream recording

Happy streaming! 🎬🚀
