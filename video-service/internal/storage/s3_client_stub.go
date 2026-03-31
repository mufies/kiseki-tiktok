// +build !aws

package storage

// S3StorageClient is not available without the 'aws' build tag
// To use S3, install dependencies and build with: go build -tags aws
// For most use cases, use MinioStorageClient instead
