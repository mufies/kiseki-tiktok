package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StreamStatus string

const (
	StreamStatusOffline StreamStatus = "offline"
	StreamStatusLive    StreamStatus = "live"
	StreamStatusEnding  StreamStatus = "ending"
)

type Stream struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	StreamKey    string       `gorm:"type:varchar(64);unique;not null" json:"stream_key"`
	Title        string       `gorm:"type:varchar(255)" json:"title"`
	Description  string       `gorm:"type:text" json:"description"`
	ThumbnailURL string       `gorm:"type:varchar(512)" json:"thumbnail_url,omitempty"`
	Status       StreamStatus `gorm:"type:varchar(20);default:'offline';index" json:"status"`
	ViewerCount  int64        `gorm:"default:0" json:"viewer_count"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	EndedAt      *time.Time   `json:"ended_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	SaveVOD      bool         `gorm:"default:true" json:"save_vod"`
}

func (Stream) TableName() string {
	return "streams"
}

func (s *Stream) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	if s.StreamKey == "" {
		s.StreamKey = generateStreamKey()
	}

	return nil
}

func generateStreamKey() string {
	key := uuid.New().String()
	return "sk_" + key[:8] + key[9:13] + key[14:18] + key[19:23] + key[24:]
}

func (s *Stream) IsLive() bool {
	return s.Status == StreamStatusLive
}

func (s *Stream) CanGoLive() bool {
	return s.Status == StreamStatusOffline
}
