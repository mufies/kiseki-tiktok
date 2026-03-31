# Video Service Refactoring Guide

## Overview

This refactoring abstracts the storage layer and applies SOLID principles to make the video service testable, maintainable, and extensible.

## SOLID Principles Applied

### 1. Single Responsibility Principle (SRP)
Each component has one reason to change:
- `VideoService`: Business logic only
- `StorageClient`: File operations only
- `AuthorizationService`: Permission checks only
- `VideoRepository`: Data persistence only

### 2. Open/Closed Principle (OCP)
- Authorization rules are extensible without modifying existing code
- New storage backends can be added without changing VideoService

### 3. Liskov Substitution Principle (LSP)
- Any `StorageClient` implementation (Minio, S3, Mock) can be swapped
- Any `AuthorizationRule` can be used interchangeably

### 4. Interface Segregation Principle (ISP)
- Clean, focused interfaces with only necessary methods
- Clients don't depend on interfaces they don't use

### 5. Dependency Inversion Principle (DIP)
- VideoService depends on abstractions (interfaces), not concrete implementations
- Storage and authorization details are injected

## Architecture

```
┌─────────────────────────────────────────────┐
│           VideoServiceRefactored            │
│         (Business Logic Only)               │
└──────┬─────────────┬─────────────┬──────────┘
       │             │             │
       ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────────┐
│Repository│  │ Storage  │  │Authorization │
│Interface │  │Interface │  │   Service    │
└────┬─────┘  └────┬─────┘  └──────┬───────┘
     │             │                │
     ▼             ▼                ▼
┌──────────┐  ┌──────────┐  ┌──────────────┐
│  Gorm    │  │  Minio   │  │ Ownership    │
│  Impl    │  │   S3     │  │    Rules     │
│          │  │  Mock    │  │              │
└──────────┘  └──────────┘  └──────────────┘
```

## Key Components

### Storage Abstraction

```go
type StorageClient interface {
    PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
    GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
    RemoveObject(ctx context.Context, bucket, key string) error
    PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (*url.URL, error)
}
```

**Implementations:**
- `MinioStorageClient`: Production Minio adapter
- `S3StorageClient`: AWS S3 adapter
- `MockStorageClient`: In-memory testing implementation

### Authorization System

```go
type AuthorizationRule interface {
    Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error
}
```

**Built-in Rules:**
- `OwnershipRule`: Only owner can modify
- `PublicAccessRule`: Anyone can view
- `AdminRule`: Admin bypass
- `CompositeRule`: Combine multiple rules with OR logic

### Repository Interface

```go
type VideoRepositoryInterface interface {
    Save(video *model.Video) error
    FindByID(id uuid.UUID) (*model.Video, error)
    FindByOwnerID(ownerID uuid.UUID) ([]model.Video, error)
    Update(id uuid.UUID, video *model.Video) error
    Delete(id uuid.UUID) error
}
```

## Migration Path (Zero Downtime)

### Phase 1: Introduce New Code (No Breaking Changes)

Keep old `VideoService`, introduce new `VideoServiceRefactored` side-by-side:

```go
// main.go
func main() {
    // Old service (still works)
    oldService := service.NewVideoService(repo, minioClient, bucket)

    // New service (ready for testing)
    storageClient := storage.NewMinioStorageClient(minioClient)
    authService := authorization.NewVideoAuthorizationService()
    newService := service.NewVideoServiceRefactored(repo, storageClient, authService, bucket)

    // Use old service for now
    handler := handler.NewVideoHandler(oldService)
}
```

### Phase 2: Gradual Cutover

Switch endpoints one at a time:

```go
type VideoHandler struct {
    oldService *service.VideoService
    newService *service.VideoServiceRefactored
}

func (h *VideoHandler) Upload(c *gin.Context) {
    // Use new service for uploads
    result, err := h.newService.Upload(ctx, ...)
    // ...
}

func (h *VideoHandler) GetByID(c *gin.Context) {
    // Still use old service for reads
    result, err := h.oldService.GetByID(id)
    // ...
}
```

### Phase 3: Full Migration

Once validated, switch all endpoints to new service and remove old code.

## Configuration Examples

### Production with Minio

```go
// cmd/main.go
package main

import (
    "github.com/kiseki/video-service/internal/storage"
    "github.com/kiseki/video-service/internal/authorization"
    "github.com/kiseki/video-service/internal/service"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
    // Initialize Minio client
    minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
        Secure: cfg.MinioUseSSL,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create storage adapter
    storageClient := storage.NewMinioStorageClient(minioClient)

    // Create authorization service
    authService := authorization.NewVideoAuthorizationService()

    // Create repository
    repo := repository.NewVideoRepository(db)

    // Inject dependencies
    videoService := service.NewVideoServiceRefactored(
        repo,
        storageClient,
        authService,
        cfg.MinioBucket,
    )

    // Use service in handlers
    handler := handler.NewVideoHandler(videoService)
}
```

### Production with AWS S3

```go
import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/kiseki/video-service/internal/storage"
)

func main() {
    // Initialize S3 client
    sess := session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    }))
    s3Client := s3.New(sess)

    // Create storage adapter
    storageClient := storage.NewS3StorageClient(s3Client)

    // Create authorization service
    authService := authorization.NewVideoAuthorizationService()

    // Create repository
    repo := repository.NewVideoRepository(db)

    // Inject dependencies
    videoService := service.NewVideoServiceRefactored(
        repo,
        storageClient,
        authService,
        "my-s3-bucket",
    )
}
```

### Testing with Mock Storage

```go
func TestVideoUpload(t *testing.T) {
    // Use mock implementations
    mockRepo := NewMockVideoRepository()
    mockStorage := storage.NewMockStorageClient()
    authService := authorization.NewVideoAuthorizationService()

    service := service.NewVideoServiceRefactored(
        mockRepo,
        mockStorage,
        authService,
        "test-bucket",
    )

    // Test without external dependencies
    video, err := service.Upload(ctx, ownerID, "Test", "Desc", file, header, nil, nil)
    assert.NoError(t, err)
    assert.True(t, mockStorage.ObjectExists("test-bucket", video.FileName))
}
```

### Custom Authorization Rules

```go
// Allow owner OR admins to delete
func main() {
    adminIDs := []uuid.UUID{
        uuid.MustParse("admin-1-uuid"),
        uuid.MustParse("admin-2-uuid"),
    }

    ownershipRule := authorization.NewOwnershipRule()
    adminRule := authorization.NewAdminRule(adminIDs)

    // Combine with OR logic
    deleteRule := authorization.NewCompositeRule(ownershipRule, adminRule)

    authService := authorization.NewVideoAuthorizationServiceWithRules(
        ownershipRule,            // update: owner only
        deleteRule,               // delete: owner OR admin
        authorization.NewPublicAccessRule(), // view: anyone
    )

    videoService := service.NewVideoServiceRefactored(
        repo,
        storageClient,
        authService,
        bucket,
    )
}
```

## Environment Variables

No changes required! Existing `.env` configuration works:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=video_service

# Storage (Minio or S3)
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=videos
MINIO_USE_SSL=false

# For S3, use:
# AWS_REGION=us-east-1
# AWS_S3_BUCKET=my-videos-bucket
```

## Testing Strategy

### Unit Tests (Fast, Isolated)

```bash
go test ./internal/service/... -v
```

Uses mock storage and repository - no external dependencies.

### Integration Tests (With Real Storage)

```go
func TestIntegration_Upload(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Use real Minio
    minioClient := setupRealMinioClient(t)
    storageClient := storage.NewMinioStorageClient(minioClient)

    // Test with real storage
    service := service.NewVideoServiceRefactored(mockRepo, storageClient, authService, "test-bucket")
    // ...
}
```

Run: `go test ./internal/service/... -v` (skips integration tests)
Run: `go test ./internal/service/... -v -short=false` (includes integration tests)

## Performance Comparison

No performance degradation:
- Interface calls are optimized by Go compiler (inlining)
- Same number of network calls
- Negligible memory overhead

**Benchmark:**
```bash
go test -bench=. ./internal/service/...
```

## Benefits Summary

✅ **Testability**: Mock storage for fast unit tests
✅ **Flexibility**: Swap Minio ↔ S3 ↔ Custom storage
✅ **Maintainability**: Clear separation of concerns
✅ **Extensibility**: Add new authorization rules without changing core logic
✅ **Type Safety**: Compile-time verification of interface implementations
✅ **Zero Downtime**: Incremental migration path

## Common Patterns

### Custom Storage Implementation

```go
type CustomStorageClient struct {
    // Your storage implementation
}

func (c *CustomStorageClient) PutObject(...) error {
    // Your implementation
}

// Implement all StorageClient methods...

// Use it:
customStorage := &CustomStorageClient{}
service := service.NewVideoServiceRefactored(repo, customStorage, authService, bucket)
```

### Premium Content Rule

```go
type PremiumContentRule struct {
    premiumUsers map[uuid.UUID]bool
}

func (r *PremiumContentRule) Check(ctx context.Context, userID, videoOwnerID uuid.UUID) error {
    if r.premiumUsers[userID] {
        return nil
    }
    return fmt.Errorf("premium subscription required")
}
```

## Troubleshooting

**Issue**: `interface conversion` panic
**Solution**: Ensure your type implements all interface methods

**Issue**: Tests failing with storage errors
**Solution**: Use `MockStorageClient` for unit tests, real client for integration tests

**Issue**: Authorization not working
**Solution**: Verify `AuthorizationService` is injected correctly

## Next Steps

1. Run unit tests: `go test ./internal/service/...`
2. Deploy with new service alongside old service
3. Monitor metrics
4. Gradually switch endpoints
5. Remove old service code

## Questions?

See `/internal/service/video_service_test.go` for comprehensive examples.
