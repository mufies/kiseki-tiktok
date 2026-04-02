package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/kiseki/stream-service/config"
	"github.com/kiseki/stream-service/internal/kafka"
	"github.com/kiseki/stream-service/internal/model"
	"github.com/kiseki/stream-service/internal/repository"
	"github.com/kiseki/stream-service/internal/storage"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis keys
	redisKeyActiveViewers = "stream:%s:active_viewers"
	redisKeyViewerCount   = "stream:%s:viewer_count"
	redisKeyPeakViewers   = "stream:%s:peak_viewers"
	redisKeyTotalViewers  = "stream:%s:total_viewers"
	redisKeyStreamStatus  = "stream:%s:status"
)

type StreamService struct {
	repo    repository.StreamRepository
	storage storage.StorageClient
	redis   *redis.Client
	kafka   *kafka.KafkaProducer
	config  *config.Config
}

func NewStreamService(
	repo repository.StreamRepository,
	storageClient storage.StorageClient,
	redisClient *redis.Client,
	kafkaProducer *kafka.KafkaProducer,
	cfg *config.Config,
) *StreamService {
	return &StreamService{
		repo:    repo,
		storage: storageClient,
		redis:   redisClient,
		kafka:   kafkaProducer,
		config:  cfg,
	}
}

// CreateStream creates a new stream for a user
func (s *StreamService) CreateStream(ctx context.Context, userID uuid.UUID, title, description string) (*model.Stream, error) {
	stream := &model.Stream{
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      model.StreamStatusOffline,
		ViewerCount: 0,
		SaveVOD:     true,
	}

	if err := s.repo.Create(stream); err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return stream, nil
}

// GetStreamByID retrieves a stream by ID with real-time viewer count from Redis
func (s *StreamService) GetStreamByID(ctx context.Context, id uuid.UUID) (*model.Stream, error) {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("stream not found: %w", err)
	}

	// Get real-time viewer count from Redis if stream is live
	if stream.IsLive() && s.redis != nil {
		viewerCountKey := fmt.Sprintf(redisKeyViewerCount, id.String())
		count, err := s.redis.Get(ctx, viewerCountKey).Int64()
		if err == nil {
			stream.ViewerCount = count
		}
	}

	return stream, nil
}

// GetStreamByKey retrieves a stream by stream key (used for RTMP authentication)
func (s *StreamService) GetStreamByKey(ctx context.Context, streamKey string) (*model.Stream, error) {
	stream, err := s.repo.FindByStreamKey(streamKey)
	if err != nil {
		return nil, fmt.Errorf("stream not found: %w", err)
	}
	return stream, nil
}

// GetUserStreams retrieves all streams for a user
func (s *StreamService) GetUserStreams(ctx context.Context, userID uuid.UUID) ([]*model.Stream, error) {
	streams, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user streams: %w", err)
	}
	return streams, nil
}

// GetLiveStreams retrieves all currently live streams with pagination
func (s *StreamService) GetLiveStreams(ctx context.Context, limit, offset int) ([]*model.Stream, error) {
	streams, err := s.repo.FindLiveStreams(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get live streams: %w", err)
	}

	// Update viewer counts from Redis
	if s.redis != nil {
		for i, stream := range streams {
			viewerCountKey := fmt.Sprintf(redisKeyViewerCount, stream.ID.String())
			count, err := s.redis.Get(ctx, viewerCountKey).Int64()
			if err == nil {
				streams[i].ViewerCount = count
			}
		}
	}

	return streams, nil
}

// UpdateStream updates stream information
func (s *StreamService) UpdateStream(ctx context.Context, id uuid.UUID, userID uuid.UUID, title, description string) (*model.Stream, error) {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("stream not found: %w", err)
	}

	// Check ownership
	if stream.UserID != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this stream")
	}

	// Update fields
	if title != "" {
		stream.Title = title
	}
	if description != "" {
		stream.Description = description
	}

	if err := s.repo.Update(stream); err != nil {
		return nil, fmt.Errorf("failed to update stream: %w", err)
	}

	// Publish update event
	if s.kafka != nil {
		event := kafka.StreamUpdatedEvent{
			StreamID:    id,
			Title:       stream.Title,
			Description: stream.Description,
			UpdatedAt:   time.Now(),
		}
		_ = s.kafka.PublishEvent(ctx, s.config.KafkaTopics.StreamUpdate, id.String(), event)
	}

	return stream, nil
}

// StartStream starts a stream (called when RTMP connection begins or manually)
func (s *StreamService) StartStream(ctx context.Context, id uuid.UUID) error {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	if !stream.CanGoLive() {
		return fmt.Errorf("stream cannot go live: current status is %s", stream.Status)
	}

	// Update stream status
	now := time.Now()
	stream.Status = model.StreamStatusLive
	stream.StartedAt = &now

	if err := s.repo.Update(stream); err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Initialize Redis counters
	if s.redis != nil {
		streamID := id.String()
		s.redis.Set(ctx, fmt.Sprintf(redisKeyViewerCount, streamID), 0, 24*time.Hour)
		s.redis.Set(ctx, fmt.Sprintf(redisKeyPeakViewers, streamID), 0, 24*time.Hour)
		s.redis.Set(ctx, fmt.Sprintf(redisKeyTotalViewers, streamID), 0, 24*time.Hour)
		s.redis.Set(ctx, fmt.Sprintf(redisKeyStreamStatus, streamID), "live", 24*time.Hour)
	}

	// Publish stream started event
	if s.kafka != nil {
		event := kafka.StreamStartedEvent{
			StreamID:  id,
			UserID:    stream.UserID,
			Title:     stream.Title,
			StartedAt: now,
		}
		_ = s.kafka.PublishEvent(ctx, s.config.KafkaTopics.StreamStarted, id.String(), event)
	}

	return nil
}

// EndStream ends a stream (called when RTMP disconnects or manually)
func (s *StreamService) EndStream(ctx context.Context, id uuid.UUID) error {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	if !stream.IsLive() {
		return fmt.Errorf("stream is not live")
	}

	// Update stream status
	now := time.Now()
	stream.Status = model.StreamStatusOffline
	stream.EndedAt = &now

	if err := s.repo.Update(stream); err != nil {
		return fmt.Errorf("failed to end stream: %w", err)
	}

	// Calculate statistics from Redis
	var peakViewers, totalViewers int64
	var duration int64

	if s.redis != nil {
		streamID := id.String()
		peakViewers, _ = s.redis.Get(ctx, fmt.Sprintf(redisKeyPeakViewers, streamID)).Int64()
		totalViewers, _ = s.redis.Get(ctx, fmt.Sprintf(redisKeyTotalViewers, streamID)).Int64()

		// Clear Redis data
		s.redis.Del(ctx,
			fmt.Sprintf(redisKeyViewerCount, streamID),
			fmt.Sprintf(redisKeyPeakViewers, streamID),
			fmt.Sprintf(redisKeyTotalViewers, streamID),
			fmt.Sprintf(redisKeyStreamStatus, streamID),
			fmt.Sprintf(redisKeyActiveViewers, streamID),
		)
	}

	// Calculate duration
	if stream.StartedAt != nil {
		duration = int64(now.Sub(*stream.StartedAt).Seconds())
	}

	// Publish stream ended event
	if s.kafka != nil {
		event := kafka.StreamEndedEvent{
			StreamID:     id,
			UserID:       stream.UserID,
			EndedAt:      now,
			Duration:     duration,
			PeakViewers:  peakViewers,
			TotalViewers: totalViewers,
		}
		_ = s.kafka.PublishEvent(ctx, s.config.KafkaTopics.StreamEnded, id.String(), event)
	}

	// TODO: Trigger VOD processing if SaveVOD is true
	// This would involve:
	// 1. Converting HLS segments to a single video file
	// 2. Uploading to video-service
	// 3. Publishing VODReady event

	return nil
}

// JoinStream increments viewer count when a user joins
func (s *StreamService) JoinStream(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	if !stream.IsLive() {
		return fmt.Errorf("stream is not live")
	}

	streamID := id.String()

	// Increment viewer count in Redis
	if s.redis != nil {
		viewerCountKey := fmt.Sprintf(redisKeyViewerCount, streamID)
		newCount, err := s.redis.Incr(ctx, viewerCountKey).Result()
		if err != nil {
			return fmt.Errorf("failed to increment viewer count: %w", err)
		}

		// Update peak viewers if necessary
		peakKey := fmt.Sprintf(redisKeyPeakViewers, streamID)
		currentPeak, _ := s.redis.Get(ctx, peakKey).Int64()
		if newCount > currentPeak {
			s.redis.Set(ctx, peakKey, newCount, 24*time.Hour)
		}

		// Increment total unique viewers
		totalKey := fmt.Sprintf(redisKeyTotalViewers, streamID)
		s.redis.Incr(ctx, totalKey)

		// Add to active viewers set if user is authenticated
		if userID != nil {
			activeViewersKey := fmt.Sprintf(redisKeyActiveViewers, streamID)
			s.redis.SAdd(ctx, activeViewersKey, userID.String())
		}

		// Update database viewer count periodically (async)
		if newCount%10 == 0 { // Update DB every 10 viewers
			go s.repo.IncrementViewerCount(id)
		}
	} else {
		// Fallback to database if Redis unavailable
		if err := s.repo.IncrementViewerCount(id); err != nil {
			return fmt.Errorf("failed to increment viewer count: %w", err)
		}
	}

	// Publish viewer joined event
	if s.kafka != nil {
		viewerCount, _ := s.redis.Get(ctx, fmt.Sprintf(redisKeyViewerCount, streamID)).Int64()
		event := kafka.ViewerJoinedEvent{
			StreamID:    id,
			ViewerCount: viewerCount,
			JoinedAt:    time.Now(),
		}
		if userID != nil {
			event.UserID = *userID
		}
		_ = s.kafka.PublishEvent(ctx, s.config.KafkaTopics.ViewerJoined, id.String(), event)
	}

	return nil
}

// LeaveStream decrements viewer count when a user leaves
func (s *StreamService) LeaveStream(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	// Verify stream exists
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	streamID := id.String()

	// Decrement viewer count in Redis
	if s.redis != nil {
		viewerCountKey := fmt.Sprintf(redisKeyViewerCount, streamID)
		newCount, err := s.redis.Decr(ctx, viewerCountKey).Result()
		if err != nil {
			return fmt.Errorf("failed to decrement viewer count: %w", err)
		}

		// Ensure count doesn't go negative
		if newCount < 0 {
			s.redis.Set(ctx, viewerCountKey, 0, 24*time.Hour)
			newCount = 0
		}

		// Remove from active viewers set
		if userID != nil {
			activeViewersKey := fmt.Sprintf(redisKeyActiveViewers, streamID)
			s.redis.SRem(ctx, activeViewersKey, userID.String())
		}
	} else {
		// Fallback to database
		if err := s.repo.DecrementViewerCount(id); err != nil {
			return fmt.Errorf("failed to decrement viewer count: %w", err)
		}
	}

	// Publish viewer left event
	if s.kafka != nil {
		viewerCount, _ := s.redis.Get(ctx, fmt.Sprintf(redisKeyViewerCount, streamID)).Int64()
		event := kafka.ViewerLeftEvent{
			StreamID:    id,
			ViewerCount: viewerCount,
			LeftAt:      time.Now(),
		}
		if userID != nil {
			event.UserID = *userID
		}
		_ = s.kafka.PublishEvent(ctx, s.config.KafkaTopics.ViewerLeft, id.String(), event)
	}

	return nil
}

// DeleteStream deletes a stream (only if offline)
func (s *StreamService) DeleteStream(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	// Check ownership
	if stream.UserID != userID {
		return fmt.Errorf("unauthorized: user does not own this stream")
	}

	// Can only delete offline streams
	if stream.IsLive() {
		return fmt.Errorf("cannot delete live stream")
	}

	// TODO: Delete associated files from storage (HLS segments, VOD, thumbnails)

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	return nil
}

// GetPlaybackURL generates the HLS playback URL for a live stream
func (s *StreamService) GetPlaybackURL(ctx context.Context, id uuid.UUID) (string, error) {
	stream, err := s.repo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("stream not found: %w", err)
	}

	if !stream.IsLive() {
		return "", fmt.Errorf("stream is not live")
	}

	// TODO: Implement HLS transcoding in the built-in RTMP server
	// For now, return the RTMP URL that can be used with a player that supports RTMP
	playbackURL := fmt.Sprintf("rtmp://localhost:%s/live/%s", s.config.RTMPPort, stream.StreamKey)

	return playbackURL, nil
}

// GeneratePresignedURL generates a presigned URL for accessing stream files
func (s *StreamService) GeneratePresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := s.storage.PresignedGetObject(ctx, bucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// GenerateThumbnailURL generates a public URL for stream thumbnail
func (s *StreamService) GenerateThumbnailURL(ctx context.Context, stream *model.Stream) (string, error) {
	if stream.ThumbnailURL == "" {
		return "", nil
	}

	scheme := "http"
	if s.config.MinioPresignedUseSSL {
		scheme = "https"
	}

	encodedFilename := url.PathEscape(stream.ThumbnailURL)
	publicURL := fmt.Sprintf("%s://%s/%s/%s",
		scheme,
		s.config.MinioPublicEndpoint,
		s.config.MinioThumbnailsBucket,
		encodedFilename,
	)

	return publicURL, nil
}

// GetActiveViewers returns the list of active viewer IDs for a stream
func (s *StreamService) GetActiveViewers(ctx context.Context, id uuid.UUID) ([]string, error) {
	if s.redis == nil {
		return []string{}, nil
	}

	activeViewersKey := fmt.Sprintf(redisKeyActiveViewers, id.String())
	viewers, err := s.redis.SMembers(ctx, activeViewersKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active viewers: %w", err)
	}

	return viewers, nil
}
