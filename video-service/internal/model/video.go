package model

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"ownerId"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	FileName    string    `gorm:"not null" json:"-"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
