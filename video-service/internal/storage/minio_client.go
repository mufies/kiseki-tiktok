package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorageClient struct {
	client          *minio.Client
	presignedClient *minio.Client
}

func NewMinioStorageClient(client *minio.Client) *MinioStorageClient {
	return &MinioStorageClient{
		client:          client,
		presignedClient: client, // Default to same client
	}
}

func (m *MinioStorageClient) SetPresignedClient(client *minio.Client) {
	m.presignedClient = client
}

func (m *MinioStorageClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinioStorageClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *MinioStorageClient) RemoveObject(ctx context.Context, bucket, key string) error {
	return m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (m *MinioStorageClient) PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (*url.URL, error) {
	// Use presigned client for generating URLs accessible from browser
	return m.presignedClient.PresignedGetObject(ctx, bucket, key, expires, nil)
}
