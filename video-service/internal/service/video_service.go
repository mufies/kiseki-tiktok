package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/model"
	"github.com/kiseki/video-service/internal/repository"
	"github.com/minio/minio-go/v7"
)

type VideoService struct {
	repo        *repository.VideoRepository
	minioClient *minio.Client
	bucket      string
}

func NewVideoService(repo *repository.VideoRepository, mc *minio.Client, bucket string) *VideoService {
	return &VideoService{repo: repo, minioClient: mc, bucket: bucket}
}

func (s *VideoService) Upload(
	ownerID uuid.UUID,
	title string,
	description string,
	file multipart.File,
	header *multipart.FileHeader,
	categories []string,
	hashtags []string,
) (*model.Video, error) {
	filename := uuid.New().String() + "_" + header.Filename

	_, err := s.minioClient.PutObject(
		context.Background(),
		s.bucket,
		filename,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return nil, err
	}

	video := &model.Video{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		FileName:    filename,
		Hashtags:    hashtags,
		MimeType:    header.Header.Get("Content-Type"),
		Size:        header.Size,
	}
	if err := s.repo.Save(video); err != nil {
		return nil, err
	}
	return video, nil
}

func (s *VideoService) GetByID(id uuid.UUID) (*model.Video, string, time.Time, error) {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	streamURL, expiresAt, err := s.GeneratePresignedURL(video.FileName)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return video, streamURL, expiresAt, nil
}

// GeneratePresignedURL creates a new presigned URL for streaming video
// URL expires after 1 hour, returns both URL and expiration time
func (s *VideoService) GeneratePresignedURL(filename string) (string, time.Time, error) {
	expiration := time.Now().Add(1 * time.Hour)
	url, err := s.minioClient.PresignedGetObject(
		context.Background(),
		s.bucket,
		filename,
		1*time.Hour, // 1 hour validity
		nil,
	)
	if err != nil {
		return "", time.Time{}, err
	}
	return url.String(), expiration, nil
}

// GetPresignedURL retrieves video and generates fresh presigned URL with expiration time
func (s *VideoService) GetPresignedURL(id uuid.UUID) (string, time.Time, error) {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return "", time.Time{}, err
	}
	return s.GeneratePresignedURL(video.FileName)
}

// UpdateVideo updates video title and hashtags
// Only owner can update their video
func (s *VideoService) UpdateVideo(videoID uuid.UUID, ownerID uuid.UUID, title string, hashtags []string) (*model.Video, error) {
	video, err := s.repo.FindByID(videoID)
	if err != nil {
		return nil, err
	}

	// Check if user is the owner
	if video.OwnerID != ownerID {
		return nil, fmt.Errorf("unauthorized: only owner can update video")
	}

	// Update fields
	video.Title = title
	video.Hashtags = hashtags

	if err := s.repo.Update(videoID, video); err != nil {
		return nil, err
	}

	return video, nil
}

func (s *VideoService) GetByOwner(id uuid.UUID) ([]model.Video, error) {
	return s.repo.FindByOwnerID(id)
}

func (s *VideoService) Delete(id uuid.UUID, ownerID uuid.UUID) error {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if video.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: only owner can delete video")
	}

	if err := s.minioClient.RemoveObject(context.Background(), s.bucket, video.FileName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
