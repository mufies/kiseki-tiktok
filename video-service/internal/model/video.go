package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Video struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID           uuid.UUID `gorm:"type:uuid;not null" json:"ownerId"`
	Title             string    `gorm:"not null" json:"title"`
	Description       string    `json:"description"`
	FileName          string    `gorm:"not null" json:"-"`
	MimeType          string    `json:"mimeType"`
	Size              int64     `json:"size"`
	ThumbnailFileName string    `json:"-"`
	VideoThumbnail    string    `gorm:"-" json:"videoThumbnail"`
	// Recommendation metadata
	Hashtags  pq.StringArray `gorm:"type:text[]" json:"hashtags"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`

	// Extended data (not stored in DB, populated on demand)
	Owner        *VideoOwner       `gorm:"-" json:"owner,omitempty"`
	Interactions *VideoInteraction `gorm:"-" json:"interactions,omitempty"`
}

type VideoOwner struct {
	UserID          string `json:"userId"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName,omitempty"`
	ProfileImageURL string `json:"profileImageUrl,omitempty"`
	FollowersCount  int64  `json:"followersCount"`
	FollowingCount  int64  `json:"followingCount"`
	IsVerified      bool   `json:"isVerified"`
	IsFollowed      bool   `json:"isFollowed"`
}

type VideoInteraction struct {
	LikeCount     int64 `json:"likeCount"`
	CommentCount  int64 `json:"commentCount"`
	BookmarkCount int64 `json:"bookmarkCount"`
	ViewCount     int64 `json:"viewCount"`
	IsLiked       bool  `json:"isLiked"`
	IsBookmarked  bool  `json:"isBookmarked"`
}
