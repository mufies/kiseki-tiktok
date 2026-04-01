package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

// StorageClient interface for object storage operations
type StorageClient interface {
	PutObject(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, bucket, objectName string) error
	PresignedGetObject(ctx context.Context, bucket, objectName string, expiry time.Duration) (*url.URL, error)
	PresignedPutObject(ctx context.Context, bucket, objectName string, expiry time.Duration) (*url.URL, error)
	SetPresignedClient(client *minio.Client)
}

type minioStorageClient struct {
	client          *minio.Client
	presignedClient *minio.Client // Separate client for presigned URLs
}

func NewMinioStorageClient(client *minio.Client) StorageClient {
	return &minioStorageClient{
		client:          client,
		presignedClient: client,
	}
}

func (m *minioStorageClient) SetPresignedClient(client *minio.Client) {
	m.presignedClient = client
}

func (m *minioStorageClient) PutObject(
	ctx context.Context,
	bucket, objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := m.client.PutObject(
		ctx,
		bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	return err
}

func (m *minioStorageClient) GetObject(
	ctx context.Context,
	bucket, objectName string,
) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

func (m *minioStorageClient) RemoveObject(
	ctx context.Context,
	bucket, objectName string,
) error {
	return m.client.RemoveObject(
		ctx,
		bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
}

func (m *minioStorageClient) PresignedGetObject(
	ctx context.Context,
	bucket, objectName string,
	expiry time.Duration,
) (*url.URL, error) {
	return m.presignedClient.PresignedGetObject(
		ctx,
		bucket,
		objectName,
		expiry,
		url.Values{},
	)
}

func (m *minioStorageClient) PresignedPutObject(
	ctx context.Context,
	bucket, objectName string,
	expiry time.Duration,
) (*url.URL, error) {
	return m.presignedClient.PresignedPutObject(ctx, bucket, objectName, expiry)
}
