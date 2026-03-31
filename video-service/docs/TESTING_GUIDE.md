# Testing Guide for Refactored Video Service

## Overview

The refactored architecture enables comprehensive testing at multiple levels without external dependencies.

## Test Structure

```
internal/
├── service/
│   ├── video_service_refactored.go
│   ├── video_service_test.go         # Unit tests with mocks
│   └── video_service_integration_test.go  # Integration tests (optional)
├── storage/
│   ├── mock_client.go                # Mock storage for testing
│   └── storage_test.go
└── authorization/
    ├── rules_test.go
    └── video_authorization_service_test.go
```

## Running Tests

### All Unit Tests (Fast)
```bash
go test ./... -v
```

### Specific Package
```bash
go test ./internal/service -v
```

### With Coverage
```bash
go test ./internal/service -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests Only
```bash
go test ./... -v -short=false -run Integration
```

## Unit Testing Examples

### Test Upload Success

```go
func TestUpload_Success(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := NewMockVideoRepository()
    mockStorage := storage.NewMockStorageClient()
    authService := authorization.NewVideoAuthorizationService()

    service := service.NewVideoServiceRefactored(
        mockRepo,
        mockStorage,
        authService,
        "test-bucket",
    )

    fileContent := []byte("video data")
    file := bytes.NewReader(fileContent)
    header := createMockFileHeader("test.mp4", int64(len(fileContent)))

    // Act
    video, err := service.Upload(
        ctx,
        uuid.New(),
        "My Video",
        "Description",
        file,
        header,
        nil,
        []string{"#test"},
    )

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, video)
    assert.Equal(t, "My Video", video.Title)
    assert.True(t, mockStorage.ObjectExists("test-bucket", video.FileName))
}
```

### Test Storage Failure Handling

```go
func TestUpload_StorageFailure(t *testing.T) {
    mockStorage := storage.NewMockStorageClient()
    mockStorage.SetError("put", errors.New("storage unavailable"))

    service := service.NewVideoServiceRefactored(
        NewMockVideoRepository(),
        mockStorage,
        authorization.NewVideoAuthorizationService(),
        "test-bucket",
    )

    _, err := service.Upload(ctx, ownerID, "Title", "Desc", file, header, nil, nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "storage")
}
```

### Test Authorization

```go
func TestUpdateVideo_Unauthorized(t *testing.T) {
    mockRepo := NewMockVideoRepository()
    video := &model.Video{
        ID:      uuid.New(),
        OwnerID: uuid.New(),
        Title:   "Original",
    }
    mockRepo.Save(video)

    service := service.NewVideoServiceRefactored(
        mockRepo,
        storage.NewMockStorageClient(),
        authorization.NewVideoAuthorizationService(),
        "test-bucket",
    )

    // Try to update with different user
    differentUser := uuid.New()
    _, err := service.UpdateVideo(ctx, video.ID, differentUser, "Hacked", nil)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unauthorized")
}
```

## Mock Implementations

### MockVideoRepository

Already provided in `video_service_test.go`:

```go
type MockVideoRepository struct {
    videos map[uuid.UUID]*model.Video
    err    error
}

// Configure to return errors
mockRepo.err = errors.New("database error")
```

### MockStorageClient

Features:
- In-memory storage
- Error simulation
- Test helpers

```go
mockStorage := storage.NewMockStorageClient()

// Simulate storage failure
mockStorage.SetError("put", errors.New("disk full"))

// Verify operations
assert.True(t, mockStorage.ObjectExists("bucket", "key"))
assert.Equal(t, 3, mockStorage.GetObjectCount())
```

## Integration Testing

### Setup

Create `video_service_integration_test.go`:

```go
// +build integration

package service

import (
    "testing"
    "github.com/minio/minio-go/v7"
)

func setupRealMinio(t *testing.T) *minio.Client {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client, err := minio.New("localhost:9000", &minio.Options{
        Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
        Secure: false,
    })
    if err != nil {
        t.Fatal(err)
    }
    return client
}

func TestIntegration_FullUploadFlow(t *testing.T) {
    minioClient := setupRealMinio(t)
    storageClient := storage.NewMinioStorageClient(minioClient)

    // Use real storage, mock DB
    service := service.NewVideoServiceRefactored(
        NewMockVideoRepository(),
        storageClient,
        authorization.NewVideoAuthorizationService(),
        "integration-test-bucket",
    )

    // Test actual file upload
    video, err := service.Upload(...)
    assert.NoError(t, err)

    // Verify file exists in real Minio
    _, err = minioClient.StatObject(ctx, "integration-test-bucket", video.FileName, minio.StatObjectOptions{})
    assert.NoError(t, err)

    // Cleanup
    defer service.Delete(ctx, video.ID, video.OwnerID)
}
```

Run integration tests:
```bash
go test -tags=integration ./internal/service -v
```

## Test Helpers

```go
// helpers_test.go
package service

import (
    "mime/multipart"
    "net/textproto"
)

func createMockFileHeader(filename string, size int64) *multipart.FileHeader {
    return &multipart.FileHeader{
        Filename: filename,
        Size:     size,
        Header: textproto.MIMEHeader{
            "Content-Type": []string{"video/mp4"},
        },
    }
}

func createTestVideo(ownerID uuid.UUID, title string) *model.Video {
    return &model.Video{
        ID:       uuid.New(),
        OwnerID:  ownerID,
        Title:    title,
        FileName: "test-file.mp4",
        Hashtags: pq.StringArray{"#test"},
    }
}
```

## Table-Driven Tests

```go
func TestUpdateVideo_Scenarios(t *testing.T) {
    tests := []struct {
        name        string
        videoOwner  uuid.UUID
        requester   uuid.UUID
        expectError bool
    }{
        {
            name:        "Owner can update",
            videoOwner:  uuid.MustParse("same-uuid"),
            requester:   uuid.MustParse("same-uuid"),
            expectError: false,
        },
        {
            name:        "Non-owner cannot update",
            videoOwner:  uuid.MustParse("owner-uuid"),
            requester:   uuid.MustParse("other-uuid"),
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := NewMockVideoRepository()
            video := createTestVideo(tt.videoOwner, "Original")
            mockRepo.Save(video)

            service := service.NewVideoServiceRefactored(
                mockRepo,
                storage.NewMockStorageClient(),
                authorization.NewVideoAuthorizationService(),
                "test-bucket",
            )

            _, err := service.UpdateVideo(ctx, video.ID, tt.requester, "New Title", nil)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Benchmarks

```go
func BenchmarkUpload(b *testing.B) {
    service := service.NewVideoServiceRefactored(
        NewMockVideoRepository(),
        storage.NewMockStorageClient(),
        authorization.NewVideoAuthorizationService(),
        "bench-bucket",
    )

    fileContent := make([]byte, 1024*1024) // 1MB
    file := bytes.NewReader(fileContent)
    header := createMockFileHeader("bench.mp4", int64(len(fileContent)))

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        file.Seek(0, 0)
        _, _ = service.Upload(context.Background(), uuid.New(), "Bench", "", file, header, nil, nil)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./internal/service
```

## CI/CD Pipeline Example

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21

      - name: Run unit tests
        run: go test ./... -v -coverprofile=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v2
        with:
          files: ./coverage.out

  integration:
    runs-on: ubuntu-latest
    services:
      minio:
        image: minio/minio
        ports:
          - 9000:9000
        env:
          MINIO_ROOT_USER: minioadmin
          MINIO_ROOT_PASSWORD: minioadmin

    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21

      - name: Run integration tests
        run: go test -tags=integration ./... -v
        env:
          MINIO_ENDPOINT: localhost:9000
```

## Test Coverage Goals

- **Unit Tests**: >80% coverage
- **Integration Tests**: Critical paths (upload, delete)
- **Authorization**: 100% coverage on rules
- **Storage Adapters**: All interface methods tested

Check coverage:
```bash
go test ./internal/service -cover
```

## Best Practices

1. **Use mocks for unit tests** - Fast, isolated, repeatable
2. **Use real services for integration tests** - Validate actual behavior
3. **Test error paths** - Don't just test happy paths
4. **Verify cleanup** - Ensure resources are released
5. **Use table-driven tests** - Cover multiple scenarios efficiently
6. **Write helpers** - Reduce test boilerplate
7. **Name tests clearly** - `TestFunction_Scenario_ExpectedOutcome`

## Common Assertions

```go
// Success case
assert.NoError(t, err)
assert.NotNil(t, result)

// Failure case
assert.Error(t, err)
assert.Contains(t, err.Error(), "expected message")

// Storage verification
assert.True(t, mockStorage.ObjectExists("bucket", "key"))
assert.Equal(t, 0, mockStorage.GetObjectCount())

// Database verification
savedVideo, err := mockRepo.FindByID(video.ID)
assert.NoError(t, err)
assert.Equal(t, "Expected Title", savedVideo.Title)
```

## Next Steps

1. Run existing tests: `go test ./internal/service -v`
2. Check coverage: `go test ./internal/service -cover`
3. Add integration tests for your specific use cases
4. Set up CI/CD pipeline
5. Monitor test execution time and optimize if needed
