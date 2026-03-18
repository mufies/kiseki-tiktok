package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Video struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"ownerId"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	FileName    string    `gorm:"not null" json:"-"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	// Recommendation metadata
	Hashtags  pq.StringArray `gorm:"type:text[]" json:"hashtags"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
