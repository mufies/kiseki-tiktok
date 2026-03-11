# Notification Service Implementation - Complete Guide

## Overview
This document describes the complete implementation of the Notification Service for the TikTok-clone microservices system, including Kafka infrastructure, real-time notifications via SignalR, and integration with existing services.

## Architecture

### Components
1. **Kafka + Zookeeper**: Message broker for event-driven notifications
2. **PostgreSQL**: Database for notification persistence
3. **Redis**: Cache for unread notification counts
4. **Notification Service (.NET 8)**: Core notification service with REST, gRPC, and SignalR
5. **Producer Services**: Interaction Service and User Service sending events to Kafka

## Infrastructure Setup

### Docker Compose Services

#### Zookeeper (Port 2181)
- Manages Kafka cluster metadata
- Persistent volumes: `zookeeper-data`, `zookeeper-logs`

#### Kafka (Ports 9092, 29092)
- Message broker with 4 topics:
  - `interaction.liked`: Like notifications
  - `interaction.commented`: Comment notifications
  - `interaction.bookmarked`: Bookmark notifications
  - `user.followed`: Follow notifications
- Persistent volume: `kafka-data`

#### PostgreSQL (Port 5432)
- 5 Databases: userdb, videodb, eventdb, interactiondb, notificationdb
- Credentials: admin/admin123
- Auto-initialized via `/init-databases.sql`

#### Redis (Port 6379)
- Caching layer for notification unread counts
- Persistent volume: `redis-data`
- AOF persistence enabled

### Service Ports

| Service | HTTP | gRPC |
|---------|------|------|
| User Service | 8081 | 9090 |
| Video Service | 8082 | 9091 |
| Event Service | 8083 | 9092 |
| Interaction Service | 8084 | - |
| **Notification Service** | **8085** | **9093** |
| API Gateway | 8080 | - |

## Notification Service Details

### Database Schema

**Table: notifications**
```sql
- Id (UUID, PK)
- UserId (String, Indexed)
- FromUserId (String)
- Type (Enum: Like, Comment, Follow, Bookmark)
- VideoId (String, Nullable)
- CommentId (String, Nullable)
- IsRead (Boolean, Default: false)
- CreatedAt (DateTime)
```

**Indexes:**
- `IX_Notifications_UserId`
- `IX_Notifications_UserId_IsRead` (composite)
- `IX_Notifications_CreatedAt`
- `IX_Notifications_UserId_IsRead_CreatedAt` (composite, for optimized pagination)

### Redis Cache Keys

- Pattern: `notification:unread:{userId}`
- Value: Integer count of unread notifications
- Strategy:
  - Increment on create
  - Delete on mark-as-read
  - Set to 0 on mark-all-read
  - Lazy load from DB on cache miss

### API Endpoints

#### REST API (Port 8085)
```
GET    /api/notifications/{userId}?page=1&pageSize=20
GET    /api/notifications/{userId}/unread-count
POST   /api/notifications/{userId}/mark-read
       Body: { "notificationIds": ["guid1", "guid2"] }
POST   /api/notifications/{userId}/mark-all-read
GET    /health
```

#### gRPC (Port 9093)
- `GetNotifications(GetNotificationsRequest) → GetNotificationsResponse`
- `MarkAsRead(MarkAsReadRequest) → MarkAsReadResponse`
- `GetUnreadCount(GetUnreadCountRequest) → GetUnreadCountResponse`

#### SignalR Hub
- WebSocket endpoint: `ws://localhost:8085/hubs/notification?userId={userId}`
- Event: `ReceiveNotification` (sent when new notification created)
- Client groups: `user:{userId}`

### Background Services

**KafkaConsumerService**
- Subscribes to 4 Kafka topics
- Manual commit after successful processing
- Error handling with logging (no dead-letter queue in dev)
- Flow:
  1. Consume message from Kafka
  2. Deserialize to `NotificationEventDto`
  3. Create `Notification` entity
  4. Save to PostgreSQL
  5. Update Redis cache
  6. Send real-time notification via SignalR
  7. Send email for FOLLOW events (placeholder)
  8. Commit offset

### Configuration

**appsettings.json**
```json
{
  "ConnectionStrings": {
    "DefaultConnection": "Host=localhost;Port=5432;Database=notificationdb;Username=admin;Password=admin123",
    "RedisConnection": "localhost:6379"
  },
  "Kafka": {
    "BootstrapServers": "localhost:29092",
    "GroupId": "notification-service-group",
    "Topics": ["interaction.liked", "interaction.commented", "interaction.bookmarked", "user.followed"]
  },
  "Smtp": {
    "Host": "smtp.gmail.com",
    "Port": 587,
    "Username": "your-email@gmail.com",
    "Password": "your-app-password",
    "FromEmail": "noreply@tiktokclone.com"
  },
  "Kestrel": {
    "Endpoints": {
      "Http": { "Url": "http://localhost:8085", "Protocols": "Http1" },
      "Grpc": { "Url": "http://localhost:9093", "Protocols": "Http2" }
    }
  }
}
```

## Producer Service Modifications

### Interaction Service

**Files Modified:**
1. `pom.xml`: Added `spring-kafka` dependency
2. `application.yml`: Added Kafka producer config
3. Created `NotificationEvent.java`: DTO for Kafka messages
4. Created `KafkaProducerService.java`: Sends like, comment, bookmark events
5. Created `VideoGrpcClient.java`: Fetches video owner via gRPC
6. Modified `InteractionService.java`:
   - Inject `KafkaProducerService` and `VideoGrpcClient`
   - `toggleLike()`: Send event only on "like" (not "unlike"), skip if self-like
   - `addComment()`: Send event after saving comment, skip if self-comment
   - `toggleBookMarked()`: Send event only on "bookmark", skip if self-bookmark

**Event Structure:**
```json
{
  "type": "LIKE|COMMENT|BOOKMARK",
  "fromUserId": "uuid-of-user-performing-action",
  "toUserId": "uuid-of-notification-recipient",
  "videoId": "uuid-of-video",
  "commentId": "uuid-of-comment" // only for COMMENT type
}
```

### User Service

**Files Modified:**
1. `application.yaml`: Added Kafka producer config (already had spring-kafka in pom.xml)
2. Created `FollowEvent.java`: DTO for follow events
3. Created `KafkaProducerService.java`: Sends follow events
4. Modified `UserService.java`:
   - Inject `KafkaProducerService`
   - `followUser()`: Send event after saving follow relationship

**Event Structure:**
```json
{
  "type": "FOLLOW",
  "fromUserId": "uuid-of-follower",
  "toUserId": "uuid-of-followed-user"
}
```

## Testing Guide

### 1. Start Infrastructure
```bash
cd /home/mufies/Code/tiktok-clone
docker-compose up -d postgres redis kafka zookeeper
```

Wait for services to be healthy:
```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

### 2. Verify Kafka Topics
```bash
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

Expected topics:
- interaction.liked
- interaction.commented
- interaction.bookmarked
- user.followed

### 3. Start Notification Service
```bash
cd notification-service/NotificationService
dotnet restore
dotnet run
```

Check logs for:
- ✅ Database migrated successfully
- ✅ Kafka consumer started
- ✅ Subscribed to topics

### 4. Test End-to-End Flow

#### Like Notification
```bash
# User A likes User B's video
curl -X POST http://localhost:8084/api/interactions/videos/{videoId}/like \
  -H "Authorization: Bearer {userA-token}"

# Check Kafka consumer logs (should show "Received message from topic interaction.liked")
# Check notification created:
curl http://localhost:8085/api/notifications/{userB-id}

# Check unread count:
curl http://localhost:8085/api/notifications/{userB-id}/unread-count
```

#### Comment Notification
```bash
curl -X POST http://localhost:8084/api/interactions/videos/{videoId}/comments \
  -H "Content-Type: application/json" \
  -d '{"content": "Great video!"}'
```

#### Follow Notification
```bash
curl -X POST http://localhost:8081/api/users/{userId}/follow \
  -H "Authorization: Bearer {token}"
```

#### SignalR Real-Time Test
```javascript
// Client-side code
const connection = new signalR.HubConnectionBuilder()
  .withUrl("http://localhost:8085/hubs/notification?userId={userId}")
  .build();

connection.on("ReceiveNotification", (notification) => {
  console.log("New notification:", notification);
});

await connection.start();
```

### 5. Monitor Kafka
```bash
# View consumer group status
docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 \
  --describe --group notification-service-group

# Tail messages from topic
docker exec kafka kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic interaction.liked --from-beginning
```

### 6. Test Redis Cache
```bash
docker exec redis redis-cli

# Check cache key
GET notification:unread:user-id-here

# Monitor all commands
MONITOR
```

## Health Checks

```bash
curl http://localhost:8085/health
```

Response:
```json
{
  "status": "Healthy",
  "results": {
    "postgresql": { "status": "Healthy" },
    "redis": { "status": "Healthy" },
    "kafka": { "status": "Healthy" }
  }
}
```

## Common Issues & Solutions

### Issue: Kafka consumer not receiving messages
**Solution:**
- Check Kafka is running: `docker ps | grep kafka`
- Verify topics exist: `docker exec kafka kafka-topics --list --bootstrap-server localhost:9092`
- Check producer is using correct bootstrap server (localhost:29092 from host, kafka:9092 from container)
- Check consumer group: `kafka-consumer-groups --describe --group notification-service-group`

### Issue: Database connection failed
**Solution:**
- Verify PostgreSQL is running and healthy
- Check connection string in appsettings.json (should use localhost when running outside Docker)
- Update to use `postgres:5432` when running in Docker container
- Verify database was created: `docker exec postgres psql -U admin -c '\l'`

### Issue: Redis cache not working
**Solution:**
- Verify Redis is running: `docker exec redis redis-cli ping` (should return PONG)
- Check connection string format: `localhost:6379` (no protocol prefix)
- Monitor Redis commands: `docker exec redis redis-cli MONITOR`

### Issue: gRPC client connection refused
**Solution:**
- Ensure Video Service is running on port 9091
- Check gRPC client configuration in Interaction Service application.yml
- When running in Docker, use service name: `video-service:9091`

## Production Considerations

### Dead Letter Queue
Current implementation logs deserialization errors. For production:
1. Add dead-letter topic configuration
2. Send failed messages to `notification.dlq`
3. Implement retry logic with exponential backoff

### Email Service
Current implementation is a placeholder. For production:
1. Configure real SMTP credentials
2. Fetch user email from User Service via gRPC
3. Implement email templates
4. Add email queue for async processing
5. Handle bounce/unsubscribe

### Scalability
1. Increase Kafka partitions for parallel processing
2. Run multiple Notification Service instances (same consumer group)
3. Add Redis Sentinel for HA
4. Use PostgreSQL read replicas for queries

### Monitoring
1. Add Prometheus metrics for:
   - Kafka consumer lag
   - Notification processing time
   - Cache hit/miss ratio
   - SignalR connection count
2. Set up alerts for consumer lag > 1000 messages
3. Log aggregation with ELK/Loki

## File Structure Summary

```
tiktok-clone/
├── docker-compose.yml (✅ Created)
├── init-databases.sql (✅ Created)
├── proto/
│   └── notification.proto (✅ Created)
├── notification-service/
│   └── NotificationService/
│       ├── Models/ (✅ 2 files)
│       ├── Data/ (✅ 1 file)
│       ├── DTOs/ (✅ 3 files)
│       ├── Configuration/ (✅ 2 files)
│       ├── Repositories/ (✅ 2 files)
│       ├── Services/ (✅ 4 files)
│       ├── BackgroundServices/ (✅ 1 file)
│       ├── Hubs/ (✅ 1 file)
│       ├── GrpcServices/ (✅ 1 file)
│       ├── Controllers/ (✅ 1 file)
│       ├── Program.cs (✅ Modified)
│       ├── appsettings.json (✅ Modified)
│       ├── NotificationService.csproj (✅ Modified)
│       └── Dockerfile (✅ Created)
├── interaction-service/
│   └── interactionservice/
│       ├── pom.xml (✅ Modified)
│       ├── src/main/resources/application.yml (✅ Modified)
│       └── src/main/java/com/kiseki/interaction/
│           ├── dto/NotificationEvent.java (✅ Created)
│           ├── kafka/KafkaProducerService.java (✅ Created)
│           ├── grpc/VideoGrpcClient.java (✅ Created)
│           └── service/InteractionService.java (✅ Modified)
└── user-service/
    └── userservice/
        ├── src/main/resources/application.yaml (✅ Modified)
        └── src/main/java/com/kiseki/userservice/
            ├── dto/FollowEvent.java (✅ Created)
            ├── kafka/KafkaProducerService.java (✅ Created)
            └── service/UserService.java (✅ Modified)
```

**Total: 37 files created/modified**

## Quick Start Commands

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Wait for services to be healthy (30-60 seconds)
docker ps

# 3. Run Notification Service
cd notification-service/NotificationService
dotnet restore
dotnet run

# 4. In another terminal, run Interaction Service
cd interaction-service/interactionservice
mvn clean install
mvn spring-boot:run

# 5. In another terminal, run User Service
cd user-service/userservice
mvn spring-boot:run

# 6. Test notification flow
# Like a video → Check notification created
# Follow a user → Check notification + email sent
# Comment on video → Check notification
```

## Next Steps

1. **Frontend Integration**:
   - Connect SignalR client to `/hubs/notification`
   - Display real-time notification toasts
   - Add notification bell icon with unread count
   - Implement notification dropdown

2. **Enhanced Features**:
   - Notification preferences (email on/off per type)
   - Mark notification as seen (different from read)
   - Notification grouping (e.g., "5 people liked your video")
   - Deep links to related content

3. **Testing**:
   - Unit tests for services
   - Integration tests for Kafka flow
   - Load tests for SignalR connections
   - E2E tests for complete notification flow

---

**Implementation Date**: 2026-03-11
**Status**: ✅ Complete and Ready for Testing
