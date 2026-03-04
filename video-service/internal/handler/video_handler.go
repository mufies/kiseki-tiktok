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

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Video file required"})
		return
	}

	video, err := h.svc.Upload(ownerID, title, description, file, header)
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

	video, streamURL, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video": video,
		"streamUrl": streamURL,
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
