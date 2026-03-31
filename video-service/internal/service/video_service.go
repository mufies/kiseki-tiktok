package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/authorization"
	"github.com/kiseki/video-service/internal/grpc/userpb"
	"github.com/kiseki/video-service/internal/model"
	"github.com/kiseki/video-service/internal/repository"
	"github.com/kiseki/video-service/internal/storage"
	"github.com/kiseki/video-service/internal/thumbnail"
)

type VideoService struct {
	repo                   repository.VideoRepositoryInterface
	storage                storage.StorageClient
	authService            authorization.AuthorizationService
	thumbnailGen           *thumbnail.Generator
	bucket                 string
	thumbnailsBucket       string
	minioPresignedEndpoint string // Internal endpoint used for presigned client
	minioPublicEndpoint    string // Public endpoint for browser access
	userClient             userpb.UserServiceClient
	interactionServiceURL  string
	httpClient             *http.Client
}

func NewVideoService(
	repo repository.VideoRepositoryInterface,
	storageClient storage.StorageClient,
	authService authorization.AuthorizationService,
	bucket string,
	thumbnailsBucket string,
	minioPresignedEndpoint string,
	minioPublicEndpoint string,
	userClient userpb.UserServiceClient,
	interactionServiceURL string,
) *VideoService {
	return &VideoService{
		repo:                   repo,
		storage:                storageClient,
		authService:            authService,
		thumbnailGen:           thumbnail.NewGenerator(),
		bucket:                 bucket,
		thumbnailsBucket:       thumbnailsBucket,
		minioPresignedEndpoint: minioPresignedEndpoint,
		minioPublicEndpoint:    minioPublicEndpoint,
		userClient:             userClient,
		interactionServiceURL:  interactionServiceURL,
		httpClient:             &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *VideoService) Upload(
	ctx context.Context,
	ownerID uuid.UUID,
	title string,
	description string,
	file multipart.File,
	header *multipart.FileHeader,
	categories []string,
	hashtags []string,
) (*model.Video, error) {
	filename := uuid.New().String() + "_" + header.Filename
	contentType := header.Header.Get("Content-Type")

	videoData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %w", err)
	}

	if err := s.storage.PutObject(
		ctx,
		s.bucket,
		filename,
		bytes.NewReader(videoData),
		int64(len(videoData)),
		contentType,
	); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	thumbnailFilename := ""
	thumbnailReader, thumbnailSize, err := s.thumbnailGen.GenerateFromVideo(bytes.NewReader(videoData), int64(len(videoData)))
	if err != nil {
		fmt.Printf("Warning: failed to generate thumbnail: %v\n", err)
	} else {
		thumbnailFilename = uuid.New().String() + "_thumbnail.jpg"
		if err := s.storage.PutObject(
			ctx,
			s.thumbnailsBucket,
			thumbnailFilename,
			thumbnailReader,
			thumbnailSize,
			"image/jpeg",
		); err != nil {
			fmt.Printf("Warning: failed to upload thumbnail: %v\n", err)
			thumbnailFilename = ""
		}
	}

	video := &model.Video{
		ID:                uuid.New(),
		OwnerID:           ownerID,
		Title:             title,
		Description:       description,
		FileName:          filename,
		ThumbnailFileName: thumbnailFilename,
		Hashtags:          hashtags,
		MimeType:          contentType,
		Size:              header.Size,
	}

	if thumbnailFilename != "" {
		thumbnailURL, _, err := s.GeneratePresignedURL(ctx, s.thumbnailsBucket, thumbnailFilename)
		if err == nil {
			video.VideoThumbnail = thumbnailURL
		}
	}

	if err := s.repo.Save(video); err != nil {
		_ = s.storage.RemoveObject(ctx, s.bucket, filename)
		if thumbnailFilename != "" {
			_ = s.storage.RemoveObject(ctx, s.thumbnailsBucket, thumbnailFilename)
		}
		return nil, fmt.Errorf("failed to save video metadata: %w", err)
	}

	return video, nil
}

func (s *VideoService) GetByID(ctx context.Context, id uuid.UUID, currentUserID *uuid.UUID) (*model.Video, string, time.Time, error) {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("video not found: %w", err)
	}

	streamURL, expiresAt, err := s.GeneratePresignedURL(ctx, s.bucket, video.FileName)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to generate URL: %w", err)
	}

	if video.ThumbnailFileName != "" {
		thumbnailURL, err := s.GenerateThumbnailURL(ctx, video)
		if err == nil {
			video.VideoThumbnail = thumbnailURL
		}
	}

	// Populate owner info
	if s.userClient != nil {
		owner, err := s.fetchOwnerInfo(ctx, video.OwnerID, currentUserID)
		if err == nil {
			video.Owner = owner
		} else {
			fmt.Printf("Warning: Failed to fetch owner info: %v\n", err)
		}
	}

	// Populate interactions
	if currentUserID != nil && s.interactionServiceURL != "" {
		interactions, err := s.fetchInteractions(ctx, id, currentUserID)
		if err == nil {
			video.Interactions = interactions
		} else {
			fmt.Printf("Warning: Failed to fetch interactions: %v\n", err)
		}
	} else if currentUserID == nil {
		fmt.Printf("Debug: currentUserID is nil, skipping interactions fetch\n")
	}

	return video, streamURL, expiresAt, nil
}

func (s *VideoService) fetchOwnerInfo(ctx context.Context, ownerID uuid.UUID, currentUserID *uuid.UUID) (*model.VideoOwner, error) {
	// Get user info
	userResp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{
		UserId: ownerID.String(),
	})
	if err != nil {
		return nil, err
	}

	owner := &model.VideoOwner{
		UserID:          userResp.User.UserId,
		Username:        userResp.User.Username,
		DisplayName:     userResp.User.DisplayName,
		ProfileImageURL: userResp.User.ProfileImageUrl,
		FollowersCount:  userResp.User.FollowersCount,
		FollowingCount:  userResp.User.FollowingCount,
		IsVerified:      false, // Not in proto yet
		IsFollowed:      false,
	}

	// Check follow status if current user is provided
	if currentUserID != nil && *currentUserID != ownerID {
		followResp, err := s.userClient.CheckFollowStatus(ctx, &userpb.UserFollowStatusRequest{
			UserId:      currentUserID.String(),
			UserIdCheck: ownerID.String(),
		})
		if err == nil {
			owner.IsFollowed = followResp.Followed
		}
	}

	return owner, nil
}

type InteractionResponse struct {
	VideoID       string `json:"videoId"`
	LikeCount     int64  `json:"likeCount"`
	CommentCount  int64  `json:"commentCount"`
	BookmarkCount int64  `json:"bookmarkCount"`
	ViewCount     int64  `json:"viewCount"`
	IsLiked       bool   `json:"isLiked"`
	IsBookmarked  bool   `json:"isBookmarked"`
}

func (s *VideoService) fetchInteractions(ctx context.Context, videoID uuid.UUID, userID *uuid.UUID) (*model.VideoInteraction, error) {
	reqURL := fmt.Sprintf("%s/interactions/videos/bulk?videoIds=%s", s.interactionServiceURL, videoID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	if userID != nil {
		req.Header.Set("X-User-Id", userID.String())
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("interaction service returned status %d", resp.StatusCode)
	}

	var interactions []InteractionResponse
	if err := json.NewDecoder(resp.Body).Decode(&interactions); err != nil {
		return nil, err
	}

	if len(interactions) == 0 {
		return &model.VideoInteraction{}, nil
	}

	interaction := interactions[0]
	return &model.VideoInteraction{
		LikeCount:     interaction.LikeCount,
		CommentCount:  interaction.CommentCount,
		BookmarkCount: interaction.BookmarkCount,
		ViewCount:     interaction.ViewCount,
		IsLiked:       interaction.IsLiked,
		IsBookmarked:  interaction.IsBookmarked,
	}, nil
}

func (s *VideoService) GeneratePresignedURL(ctx context.Context, bucket string, filename string) (string, time.Time, error) {
	expiration := time.Now().Add(1 * time.Hour)

	// For public buckets, generate direct URL (simpler, works for development)
	// For private buckets with presigned URLs, need proper network setup or proxy
	if s.minioPublicEndpoint != "" {
		// Generate direct public URL: http://localhost:9010/bucket/filename
		// URL encode the filename to handle spaces and special characters
		scheme := "http"
		encodedFilename := url.PathEscape(filename)
		publicURL := scheme + "://" + s.minioPublicEndpoint + "/" + bucket + "/" + encodedFilename
		return publicURL, expiration, nil
	}

	// Fallback: generate presigned URL if no public endpoint configured
	presignedURL, err := s.storage.PresignedGetObject(
		ctx,
		bucket,
		filename,
		1*time.Hour,
	)
	if err != nil {
		return "", time.Time{}, err
	}

	return presignedURL.String(), expiration, nil
}

func (s *VideoService) GenerateThumbnailURL(ctx context.Context, video *model.Video) (string, error) {
	if video.ThumbnailFileName == "" {
		return "", nil
	}

	thumbnailURL, _, err := s.GeneratePresignedURL(ctx, s.thumbnailsBucket, video.ThumbnailFileName)
	if err != nil {
		return "", fmt.Errorf("failed to generate thumbnail URL: %w", err)
	}

	return thumbnailURL, nil
}

func (s *VideoService) GetPresignedURL(ctx context.Context, id uuid.UUID) (string, time.Time, error) {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return "", time.Time{}, err
	}
	return s.GeneratePresignedURL(ctx, s.bucket, video.FileName)
}

func (s *VideoService) UpdateVideo(
	ctx context.Context,
	videoID uuid.UUID,
	ownerID uuid.UUID,
	title string,
	hashtags []string,
) (*model.Video, error) {
	video, err := s.repo.FindByID(videoID)
	if err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	if err := s.authService.CanUpdate(ctx, ownerID, video.OwnerID); err != nil {
		return nil, err
	}

	video.Title = title
	video.Hashtags = hashtags

	if err := s.repo.Update(videoID, video); err != nil {
		return nil, fmt.Errorf("failed to update video: %w", err)
	}

	return video, nil
}

func (s *VideoService) GetByOwner(ctx context.Context, id uuid.UUID) ([]model.Video, error) {
	videos, err := s.repo.FindByOwnerID(id)
	if err != nil {
		return nil, err
	}

	for i := range videos {
		if videos[i].ThumbnailFileName != "" {
			thumbnailURL, err := s.GenerateThumbnailURL(ctx, &videos[i])
			if err == nil {
				videos[i].VideoThumbnail = thumbnailURL
			}
		}
	}

	return videos, nil
}

func (s *VideoService) Delete(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	video, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	if err := s.authService.CanDelete(ctx, ownerID, video.OwnerID); err != nil {
		return err
	}

	if err := s.storage.RemoveObject(ctx, s.bucket, video.FileName); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if video.ThumbnailFileName != "" {
		_ = s.storage.RemoveObject(ctx, s.thumbnailsBucket, video.ThumbnailFileName)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete video metadata: %w", err)
	}

	return nil
}
