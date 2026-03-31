package storage

import (
	"context"
	"io"
	"net/url"
	"time"
)

type StorageClient interface {
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, bucket, key string) error
	PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (*url.URL, error)
}

type StorageMetadata struct {
	Bucket      string
	Key         string
	Size        int64
	ContentType string
}
