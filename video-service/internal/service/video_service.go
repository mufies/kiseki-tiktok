package service

import (
	"context"
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
		Categories:  categories,
		Hashtags:    hashtags,
		MimeType:    header.Header.Get("Content-Type"),
		Size:        header.Size,
	}
	if err := s.repo.Save(video); err != nil {
		return nil, err
	}
	return video, nil
}

func (s *VideoService) GetByID(id uuid.UUID) (*model.Video, string, error) {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return nil, "", err
	}

	url, err := s.minioClient.PresignedGetObject(
		context.Background(),
		s.bucket,
		video.FileName,
		15*time.Minute,
		nil,
	)
	if err != nil {
		return nil, "", err
	}

	return video, url.String(), nil
}

func (s *VideoService) GetByOwner(id uuid.UUID) ([]model.Video, error) {
	return s.repo.FindByOwnerID(id)
}
