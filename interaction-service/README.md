# Interaction Service

## Overview

The Interaction Service manages user interactions with videos in the TikTok clone application. It handles likes, views, comments, bookmarks, and provides interaction analytics.

## Features

- **Like Management**: Toggle likes on videos with real-time count updates
- **Bookmark Management**: Save/unsave videos to user's bookmark collection
- **Comment System**: Add and retrieve comments with user metadata enrichment
- **View Tracking**: Record unique video views per user
- **Bulk Operations**: Efficient batch fetching of interaction data for multiple videos
- **Liked Videos List**: Retrieve all videos a user has liked with metadata
- **Notification Integration**: Kafka-based event publishing for likes, comments, and bookmarks
- **Data Validation**: User and video existence validation via gRPC

## Architecture

### Technology Stack

- **Framework**: Spring Boot 3.x
- **Database**: PostgreSQL (via JPA/Hibernate)
- **Communication**:
  - REST API for client applications
  - gRPC for inter-service communication
  - Kafka for event streaming
- **Build Tool**: Maven

### Design Principles

The service follows SOLID principles:

- **Single Responsibility**: Each class has one clear purpose
  - Controller: HTTP request/response handling
  - Service: Business logic orchestration
  - Repository: Data access
  - Clients: External service communication

- **Dependency Inversion**: Depends on abstractions (interfaces) not concrete implementations
  - `VideoMetadataClient` interface with `VideoGrpcMetadataClient` implementation
  - Allows easy mocking for tests and future caching layers

- **Open/Closed**: Extensible without modification
  - Can add caching, filtering, or new interaction types without changing core logic

## API Endpoints

### 1. Toggle Like

Toggle like status for a video (like if not liked, unlike if already liked).

```http
POST /interactions/videos/{videoId}/like
Headers:
  X-User-Id: {userId}

Response 200 OK:
{
  "videoId": "uuid",
  "liked": true,
  "count": 42
}
```

**Behavior**:
- Creates like if doesn't exist
- Removes like if already exists
- Publishes Kafka event on new like (not on unlike)
- Validates user and video existence

---

### 2. Toggle Bookmark

Save or unsave a video to bookmarks.

```http
POST /interactions/videos/{videoId}/bookmarked
Headers:
  X-User-Id: {userId}

Response 200 OK:
{
  "videoId": "uuid",
  "bookmarked": true,
  "count": 15
}
```

**Behavior**:
- Creates bookmark if doesn't exist
- Removes bookmark if already exists
- Publishes Kafka event on new bookmark
- Validates user and video existence

---

### 3. Record View

Record a user viewing a video (idempotent - only counts once per user).

```http
POST /interactions/videos/{videoId}/view
Headers:
  X-User-Id: {userId}

Response 201 Created
```

**Behavior**:
- Only creates view record if user hasn't viewed this video before
- Validates user and video existence

---

### 4. Add Comment

Add a comment to a video.

```http
POST /interactions/videos/{videoId}/comment
Headers:
  X-User-Id: {userId}
Content-Type: application/json

Request Body:
{
  "content": "Great video!"
}

Response 201 Created:
{
  "id": 1,
  "userId": "uuid",
  "videoId": "uuid",
  "content": "Great video!",
  "createdAt": "2026-03-17T10:30:00"
}
```

**Validation**:
- Content cannot be empty
- Max length: 1000 characters
- Min length: 1 character (trimmed)
- Publishes Kafka event on new comment

---

### 5. Get Likes Count

Get the total number of likes for a video.

```http
GET /interactions/videos/{videoId}/likes

Response 200 OK:
{
  "videoId": "uuid",
  "liked": false,
  "count": 42
}
```

---

### 6. Get Comments

Get all comments for a video, ordered by most recent first.

```http
GET /interactions/videos/{videoId}/comments

Response 200 OK:
[
  {
    "id": 1,
    "userId": "uuid",
    "username": "john_doe",
    "userProfileImageUrl": "https://...",
    "videoId": "uuid",
    "content": "Great video!",
    "createdAt": "2026-03-17T10:30:00"
  }
]
```

**Enrichment**:
- User information (username, profile image) fetched via gRPC

---

### 7. Get Bulk Interactions

Get interaction statistics for multiple videos in one request.

```http
GET /interactions/videos/bulk?videoIds=uuid1,uuid2,uuid3
Headers:
  X-User-Id: {userId} (optional)

Response 200 OK:
[
  {
    "videoId": "uuid1",
    "likeCount": 42,
    "commentCount": 15,
    "bookmarkCount": 8,
    "viewCount": 1523,
    "isLiked": true,      // Only if X-User-Id provided
    "isBookmarked": false  // Only if X-User-Id provided
  }
]
```

**Performance**:
- Single database query for all video counts
- Efficient aggregation using GROUP BY
- Used by Feed Service to display video cards

---

### 8. Get User's Liked Videos

Retrieve all videos that a user has liked, ordered by most recent like.

```http
GET /interactions/videos/users/{userId}/liked-videos

Response 200 OK:
[
  {
    "interactionId": 1,
    "likedAt": "2026-03-17T10:30:00",
    "videoId": "uuid",
    "title": "Amazing Dance Video",
    "hashtags": ["#dance", "#viral"],
    "categories": ["Entertainment", "Dance"],
    "isAvailable": true
  },
  {
    "interactionId": 2,
    "likedAt": "2026-03-16T14:20:00",
    "videoId": "uuid2",
    "title": "Video Unavailable",
    "hashtags": [],
    "categories": [],
    "isAvailable": false
  }
]
```

**Features**:
- Ordered by most recent like first
- Includes video metadata (title, hashtags, categories)
- Handles deleted videos gracefully (isAvailable: false)
- Efficient bulk video metadata fetching

**Implementation Details** (InteractionService.java:281-330):
1. Fetches all like interactions for user
2. Extracts video IDs from interactions
3. Bulk fetches video metadata in one call (N+1 prevention)
4. Combines interaction data + video metadata
5. Returns unified response

---

## Data Model

### Interaction Entity

```java
@Entity
@Table(name = "interactions")
public class Interaction {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private UUID userId;

    @Column(nullable = false)
    private UUID videoId;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private InteractionType type;

    @Column(columnDefinition = "TEXT")
    private String content;  // Used for comments

    @CreationTimestamp
    @Column(updatable = false)
    private LocalDateTime createdAt;
}
```

### Interaction Types

```java
public enum InteractionType {
    LIKE,
    VIEW,
    COMMENT,
    BOOKMARKED
}
```

### Database Indexes

Recommended indexes for performance:

```sql
CREATE INDEX idx_user_video_type ON interactions(user_id, video_id, type);
CREATE INDEX idx_video_type ON interactions(video_id, type);
CREATE INDEX idx_user_type_created ON interactions(user_id, type, created_at DESC);
```

---

## gRPC Integration

### Client Dependencies

The service communicates with other services via gRPC:

1. **User Service** (`UserGrpcClient`)
   - `isUserExists(userId)`: Validate user existence
   - `getUserById(userId)`: Fetch user profile data

2. **Video Service** (`VideoGrpcClient`, `VideoMetadataClient`)
   - `validateVideoExists(videoId)`: Validate video existence
   - `getVideoOwnerId(videoId)`: Get video owner for notifications
   - `getVideoById(videoId)`: Fetch video metadata
   - `getBulkVideos(videoIds)`: Batch fetch video metadata

### gRPC Server

The service also exposes gRPC endpoints (InteractionGrpcService.java):

```protobuf
service InteractionService {
  rpc GetBulkInteractions(GetBulkInteractionsRequest) returns (GetBulkInteractionsResponse);
}
```

Used by other services (e.g., Feed Service) to fetch interaction data.

---

## Kafka Integration

### Event Publishing

The service publishes events to Kafka for notification service consumption:

**KafkaProducerService**:
- `sendLikeEvent(userId, videoOwnerId, videoId)`: When user likes a video
- `sendCommentEvent(userId, videoOwnerId, videoId, commentId)`: When user comments
- `sendBookmarkEvent(userId, videoOwnerId, videoId)`: When user bookmarks a video

**Event Flow**:
```
User likes video
  → InteractionService.toggleLike()
  → Save to database
  → KafkaProducerService.sendLikeEvent()
  → Notification Service consumes event
  → Push notification to video owner
```

**Business Rules**:
- Events only sent when action is performed (not on undo)
- No event if user interacts with their own content
- Events include all necessary IDs for notification service

---

## Security Considerations

### Input Validation

- **Comment Length**: Max 1000 chars to prevent DoS
- **Content Sanitization**: Trimmed to prevent whitespace-only comments
- **User/Video Validation**: All operations validate existence via gRPC

### Authentication

- User ID passed via `X-User-Id` header
- API Gateway should validate JWT and extract user ID
- Interaction Service trusts the header value from gateway

### Fraud Prevention

- Unique constraint on (userId, videoId, type) prevents duplicate interactions
- View counting is idempotent (one view per user)

---

## Configuration

### application.yml

```yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/interaction_db
    username: postgres
    password: password
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true

grpc:
  client:
    user-service:
      address: static://localhost:9091
      negotiationType: plaintext
    video-service:
      address: static://localhost:9092
      negotiationType: plaintext

kafka:
  bootstrap-servers: localhost:9093
  producer:
    key-serializer: org.apache.kafka.common.serialization.StringSerializer
    value-serializer: org.apache.kafka.common.serialization.StringSerializer

server:
  port: 8083
```

---

## Running the Service

### Prerequisites

- Java 17+
- PostgreSQL database
- Kafka broker
- User Service running on port 9091
- Video Service running on port 9092

### Build

```bash
cd interactionservice
mvn clean package
```

### Run

```bash
mvn spring-boot:run
```

### Docker

```bash
docker build -t interaction-service .
docker run -p 8083:8083 interaction-service
```

---

## Testing

### Unit Tests

Test service logic in isolation with mocked dependencies:

```java
@SpringBootTest
class InteractionServiceTest {
    @Mock
    private InteractionRepository repository;

    @Mock
    private VideoMetadataClient videoMetadataClient;

    @InjectMocks
    private InteractionService service;

    @Test
    void getUserLikedVideos_shouldReturnVideoMetadata() {
        // Test implementation
    }
}
```

### Integration Tests

Test full request/response flow with TestContainers:

```java
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@Testcontainers
class InteractionControllerIT {
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:15");

    @Test
    void toggleLike_shouldCreateAndRemoveLike() {
        // Test implementation
    }
}
```

### API Testing with cURL

```bash
# Toggle like
curl -X POST http://localhost:8083/interactions/videos/{videoId}/like \
  -H "X-User-Id: {userId}"

# Get user's liked videos
curl -X GET http://localhost:8083/interactions/videos/users/{userId}/liked-videos

# Add comment
curl -X POST http://localhost:8083/interactions/videos/{videoId}/comment \
  -H "X-User-Id: {userId}" \
  -H "Content-Type: application/json" \
  -d '{"content": "Great video!"}'
```

---

## Performance Optimizations

### Current Optimizations

1. **Bulk Video Fetching**: `getBulkVideos()` reduces N+1 queries
2. **Database Indexes**: Composite indexes on frequently queried columns
3. **Transaction Boundaries**: `@Transactional(readOnly = true)` for read operations
4. **Efficient Aggregation**: Single GROUP BY query for bulk interactions

### Future Optimizations

1. **Caching Layer**: Redis cache for video metadata
   ```java
   @Component
   public class CachedVideoMetadataClient implements VideoMetadataClient {
       private final RedisTemplate<String, VideoMetadata> redis;
       private final VideoGrpcMetadataClient delegate;
   }
   ```

2. **Bulk gRPC Endpoint**: Video Service should support `GetBulkVideos` RPC
   - Current: N sequential calls in loop
   - Future: 1 batch call with streaming response

3. **Pagination**: Add pagination to liked videos endpoint
   ```http
   GET /interactions/videos/users/{userId}/liked-videos?page=0&size=20
   ```

4. **Event Sourcing**: Store interaction events for analytics and audit

---

## Monitoring & Observability

### Metrics to Track

- Interaction creation rate (likes/comments/bookmarks per second)
- Average response time per endpoint
- gRPC call success/failure rates
- Kafka event publish success rates
- Database query performance

### Logging

The service uses SLF4J with Lombok's `@Slf4j`:

```java
log.info("User {} liked video {}", userId, videoId);
log.error("Failed to fetch video metadata for videoId: {}", videoId, e);
```

### Health Checks

Spring Boot Actuator endpoints:

```bash
GET /actuator/health
GET /actuator/metrics
GET /actuator/prometheus
```

---

## Troubleshooting

### Common Issues

1. **"Invalid user ID" error**
   - Cause: User Service is down or user doesn't exist
   - Solution: Check User Service health, verify user ID

2. **"Invalid video ID" error**
   - Cause: Video Service is down or video doesn't exist
   - Solution: Check Video Service health, verify video ID

3. **Kafka event not published**
   - Cause: Kafka broker unavailable
   - Solution: Check Kafka broker status, network connectivity

4. **Slow bulk interactions query**
   - Cause: Missing database indexes
   - Solution: Run index creation scripts

---

## API Documentation

### OpenAPI/Swagger

Access interactive API documentation at:

```
http://localhost:8083/swagger-ui.html
```

Generate OpenAPI spec:

```bash
mvn clean package
# OpenAPI spec available at target/openapi.json
```

---

## Contributing

### Code Style

- Follow Java naming conventions
- Use Lombok annotations to reduce boilerplate
- Add Javadoc for public methods
- Keep controllers thin (delegate to service layer)

### Adding New Interaction Types

1. Add enum value to `InteractionType.java`
2. Create DTO classes in `dto/request` and `dto/response`
3. Add service method in `InteractionService.java`
4. Add controller endpoint in `InteractionController.java`
5. Update documentation

---

## License

Copyright (c) 2026 Kiseki TikTok Clone Project
