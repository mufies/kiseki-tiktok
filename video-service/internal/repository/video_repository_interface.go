package repository

import (
	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/model"
)

type VideoRepositoryInterface interface {
	Save(video *model.Video) error
	FindByID(id uuid.UUID) (*model.Video, error)
	FindByOwnerID(ownerID uuid.UUID) ([]model.Video, error)
	FindAll(limit, offset int) ([]model.Video, error)
	Update(id uuid.UUID, video *model.Video) error
	Delete(id uuid.UUID) error
}
