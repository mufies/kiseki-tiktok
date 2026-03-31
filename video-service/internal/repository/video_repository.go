package repository

import (
	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/model"
	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

var _ VideoRepositoryInterface = (*VideoRepository)(nil)

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Save(video *model.Video) error {
	return r.db.Create(video).Error
}

func (r *VideoRepository) FindByID(id uuid.UUID) (*model.Video, error) {
	var video model.Video
	if err := r.db.First(&video, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *VideoRepository) FindByOwnerID(ownerId uuid.UUID) ([]model.Video, error) {
	var videos []model.Video
	if err := r.db.Where("owner_id = ?", ownerId).
		Order("created_at DESC").
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *VideoRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Video{}, "id = ?", id).Error
}

func (r *VideoRepository) FindAll(limit, offset int) ([]model.Video, error) {
	var videos []model.Video
	if err := r.db.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *VideoRepository) Update(id uuid.UUID, video *model.Video) error {
	return r.db.Model(&model.Video{}).
		Where("id = ?", id).
		Updates(video).Error
}
