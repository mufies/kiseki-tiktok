package grpcserver

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/grpc/interactionclient"
	"github.com/kiseki/video-service/internal/grpc/userpb"
	"github.com/kiseki/video-service/internal/grpc/videopb"
	"github.com/kiseki/video-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VideoGRPCServer implements videopb.VideoServiceServer
type VideoGRPCServer struct {
	videopb.UnimplementedVideoServiceServer
	repo              *repository.VideoRepository
	userClient        userpb.UserServiceClient
	interactionClient interactionclient.InteractionServiceClient
}

func NewVideoGRPCServer(repo *repository.VideoRepository, userClient userpb.UserServiceClient) *VideoGRPCServer {
	return &VideoGRPCServer{
		repo:       repo,
		userClient: userClient,
	}
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

	owner := s.getOwnerInfo(ctx, video.OwnerID.String())

	return &videopb.GetVideoResponse{
		Video: &videopb.Video{
			VideoId:      video.ID.String(),
			UserId:       video.OwnerID.String(),
			Title:        video.Title,
			Hashtags:     video.Hashtags,
			Owner:        owner,
			Description:  video.Description,
			ThumbnailUrl: video.VideoThumbnail,
			CreatedAt:    video.CreatedAt.Unix(),
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
		owner := s.getOwnerInfo(ctx, v.OwnerID.String())

		result = append(result, &videopb.Video{
			VideoId:      v.ID.String(),
			UserId:       v.OwnerID.String(),
			Title:        v.Title,
			Hashtags:     v.Hashtags,
			Owner:        owner,
			Description:  v.Description,
			ThumbnailUrl: v.VideoThumbnail,
			CreatedAt:    v.CreatedAt.Unix(),
		})
	}

	return &videopb.GetVideosResponse{Videos: result}, nil
}

// getOwnerInfo fetches user information from the user service
func (s *VideoGRPCServer) getOwnerInfo(ctx context.Context, userID string) *videopb.VideoOwner {
	if s.userClient == nil {
		return nil
	}

	userResp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("Failed to get user info for %s: %v", userID, err)
		return &videopb.VideoOwner{
			UserId:   userID,
			Username: "unknown",
		}
	}

	user := userResp.User
	return &videopb.VideoOwner{
		UserId:           user.UserId,
		Username:         user.Username,
		DisplayName:      user.DisplayName,
		ProfileImageUrl:  user.ProfileImageUrl,
		FollowersCount:   user.FollowersCount,
		FollowingCount:   user.FollowingCount,
		IsVerified:       false, // TODO: Add verified status to user model
	}
}
