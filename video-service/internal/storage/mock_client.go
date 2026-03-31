package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"
)

// MockStorageClient is an in-memory implementation for testing
type MockStorageClient struct {
	mu      sync.RWMutex
	objects map[string][]byte // key: "bucket/key"
	errors  map[string]error  // Simulate errors for specific operations
}

// NewMockStorageClient creates a new mock storage client
func NewMockStorageClient() *MockStorageClient {
	return &MockStorageClient{
		objects: make(map[string][]byte),
		errors:  make(map[string]error),
	}
}

// SetError configures the mock to return an error for a specific operation
// operation: "put", "get", "remove", "presigned"
func (m *MockStorageClient) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[operation] = err
}

// PutObject stores object in memory
func (m *MockStorageClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["put"]; err != nil {
		return err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	m.objects[fullKey] = data
	return nil
}

// GetObject retrieves object from memory
func (m *MockStorageClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errors["get"]; err != nil {
		return nil, err
	}

	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	data, exists := m.objects[fullKey]
	if !exists {
		return nil, fmt.Errorf("object not found: %s", fullKey)
	}

	return io.NopCloser(io.MultiReader(io.MultiReader(), &bytesReader{data: data})), nil
}

// RemoveObject deletes object from memory
func (m *MockStorageClient) RemoveObject(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["remove"]; err != nil {
		return err
	}

	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	delete(m.objects, fullKey)
	return nil
}

// PresignedGetObject generates mock presigned URL
func (m *MockStorageClient) PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (*url.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.errors["presigned"]; err != nil {
		return nil, err
	}

	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	if _, exists := m.objects[fullKey]; !exists {
		return nil, fmt.Errorf("object not found: %s", fullKey)
	}

	// Return mock URL
	mockURL := fmt.Sprintf("https://mock-storage.example.com/%s/%s?expires=%d",
		bucket, key, time.Now().Add(expires).Unix())
	return url.Parse(mockURL)
}

// GetObjectCount returns the number of stored objects (for testing assertions)
func (m *MockStorageClient) GetObjectCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.objects)
}

// ObjectExists checks if an object exists in storage (for testing assertions)
func (m *MockStorageClient) ObjectExists(bucket, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	_, exists := m.objects[fullKey]
	return exists
}

// bytesReader is a simple io.Reader wrapper for []byte
type bytesReader struct {
	data []byte
	pos  int
}

func (b *bytesReader) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}
