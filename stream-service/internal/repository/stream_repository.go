package repository

import (
	"github.com/google/uuid"
	"github.com/kiseki/stream-service/internal/model"
	"gorm.io/gorm"
)

type StreamRepository interface {
	Create(stream *model.Stream) error
	FindByID(id uuid.UUID) (*model.Stream, error)
	FindByStreamKey(streamKey string) (*model.Stream, error)
	FindByUserID(userID uuid.UUID) ([]*model.Stream, error)
	FindLiveStreams(limit, offset int) ([]*model.Stream, error)
	Update(stream *model.Stream) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status model.StreamStatus) error
	IncrementViewerCount(id uuid.UUID) error
	DecrementViewerCount(id uuid.UUID) error
}

type streamRepository struct {
	db *gorm.DB
}

func NewStreamRepository(db *gorm.DB) StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) Create(stream *model.Stream) error {
	return r.db.Create(stream).Error
}

func (r *streamRepository) FindByID(id uuid.UUID) (*model.Stream, error) {
	var stream model.Stream
	err := r.db.Where("id = ?", id).First(&stream).Error
	if err != nil {
		return nil, err
	}
	return &stream, nil
}

func (r *streamRepository) FindByStreamKey(streamKey string) (*model.Stream, error) {
	var stream model.Stream
	err := r.db.Where("stream_key = ?", streamKey).First(&stream).Error
	if err != nil {
		return nil, err
	}
	return &stream, nil
}

func (r *streamRepository) FindByUserID(userID uuid.UUID) ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&streams).Error
	return streams, err
}

func (r *streamRepository) FindLiveStreams(limit, offset int) ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("status = ?", model.StreamStatusLive).
		Order("viewer_count DESC").
		Limit(limit).
		Offset(offset).
		Find(&streams).Error
	return streams, err
}

func (r *streamRepository) Update(stream *model.Stream) error {
	return r.db.Save(stream).Error
}

func (r *streamRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Stream{}, "id = ?", id).Error
}

func (r *streamRepository) UpdateStatus(id uuid.UUID, status model.StreamStatus) error {
	return r.db.Model(&model.Stream{}).Where("id = ?", id).Update("status", status).Error
}

func (r *streamRepository) IncrementViewerCount(id uuid.UUID) error {
	return r.db.Model(&model.Stream{}).Where("id = ?", id).
		UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", 1)).Error
}

func (r *streamRepository) DecrementViewerCount(id uuid.UUID) error {
	return r.db.Model(&model.Stream{}).Where("id = ?", id).
		UpdateColumn("viewer_count", gorm.Expr("viewer_count - ?", 1)).
		Where("viewer_count > 0").Error
}
