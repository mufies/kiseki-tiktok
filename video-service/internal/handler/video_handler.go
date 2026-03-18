package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/service"
)

type VideoHandler struct {
	svc *service.VideoService
}

func NewVideoHandler(svc *service.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

func (h *VideoHandler) Upload(c *gin.Context) {
	ownerIDStr := c.GetHeader("X-User-Id")
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	categories := c.PostFormArray("categories")
	hashtags := c.PostFormArray("hashtags")

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Video file required"})
		return
	}

	video, err := h.svc.Upload(ownerID, title, description, file, header, categories, hashtags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, video)
}

func (h *VideoHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	video, streamURL, expiresAt, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video":     video,
		"streamUrl": streamURL,
		"expiresAt": expiresAt,
	})
}

func (h *VideoHandler) GetByOwner(c *gin.Context) {
	ownerID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	videos, err := h.svc.GetByOwner(ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, videos)
}

// GetPresignedURL generates a fresh presigned URL for streaming
// Client can call this endpoint when stream URL expires
func (h *VideoHandler) GetPresignedURL(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid video ID"})
		return
	}

	streamURL, expiresAt, err := h.svc.GetPresignedURL(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"streamUrl": streamURL,
		"expiresAt": expiresAt,
	})
}

// UpdateVideo updates video title and hashtags
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid video ID"})
		return
	}

	ownerIDStr := c.GetHeader("X-User-Id")
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	var req struct {
		Title    string   `json:"title"`
		Hashtags []string `json:"hashtags"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	video, err := h.svc.UpdateVideo(id, ownerID, req.Title, req.Hashtags)
	if err != nil {
		if err.Error() == "unauthorized: only owner can update video" {
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "Video not found"})
		}
		return
	}

	c.JSON(http.StatusOK, video)
}

func (h *VideoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid video ID"})
		return
	}

	userid, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	err = h.svc.Delete(id, userid) // Dùng = thay vì :=
	if err != nil {
		if err.Error() == "unauthorized: only owner can delete video" {
			c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "Video not found"})
			return // Thêm return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video deleted"})
}
