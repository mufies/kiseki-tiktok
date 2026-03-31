// +build aws

package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3StorageClient adapts AWS S3 client to StorageClient interface
type S3StorageClient struct {
	client *s3.S3
}

// NewS3StorageClient creates a new S3 storage adapter
func NewS3StorageClient(client *s3.S3) *S3StorageClient {
	return &S3StorageClient{client: client}
}

// PutObject uploads object to S3
func (s *S3StorageClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          aws.ReadSeekCloser(reader),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

// GetObject retrieves object from S3
func (s *S3StorageClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// RemoveObject deletes object from S3
func (s *S3StorageClient) RemoveObject(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// PresignedGetObject generates presigned URL for S3 object
func (s *S3StorageClient) PresignedGetObject(ctx context.Context, bucket, key string, expires time.Duration) (*url.URL, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expires)
	if err != nil {
		return nil, err
	}

	return url.Parse(urlStr)
}

// Helper function to convert io.Reader to io.ReadSeeker if needed
func toReadSeeker(r io.Reader) io.ReadSeeker {
	if rs, ok := r.(io.ReadSeeker); ok {
		return rs
	}
	// For testing/development - in production, ensure reader is already a ReadSeeker
	return nil
}
