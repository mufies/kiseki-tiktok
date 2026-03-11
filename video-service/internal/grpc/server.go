package grpcserver

import (
	"context"

	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/grpc/videopb"
	"github.com/kiseki/video-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VideoGRPCServer implements videopb.VideoServiceServer
type VideoGRPCServer struct {
	videopb.UnimplementedVideoServiceServer
	repo *repository.VideoRepository
}

func NewVideoGRPCServer(repo *repository.VideoRepository) *VideoGRPCServer {
	return &VideoGRPCServer{repo: repo}
}

// GetVideo returns metadata for a single video by ID.
func (s *VideoGRPCServer) GetVideo(ctx context.Context, req *videopb.GetVideoRequest) (*videopb.GetVideoResponse, error) {
	id, err := uuid.Parse(req.VideoId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid video_id: %v", err)
	}

	video, err := s.repo.FindByID(id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "video not found: %v", err)
	}

	return &videopb.GetVideoResponse{
		Video: &videopb.Video{
			VideoId:    video.ID.String(),
			Title:      video.Title,
			Categories: video.Categories,
			Hashtags:   video.Hashtags,
		},
	}, nil
}

// GetVideos returns a paginated list of videos.
func (s *VideoGRPCServer) GetVideos(ctx context.Context, req *videopb.GetVideosRequest) (*videopb.GetVideosResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	videos, err := s.repo.FindAll(limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	result := make([]*videopb.Video, 0, len(videos))
	for _, v := range videos {
		result = append(result, &videopb.Video{
			VideoId:    v.ID.String(),
			Title:      v.Title,
			Categories: v.Categories,
			Hashtags:   v.Hashtags,
		})
	}

	return &videopb.GetVideosResponse{Videos: result}, nil
}
