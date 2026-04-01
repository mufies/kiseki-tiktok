package interactionclient

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kiseki/video-service/internal/grpc/interactionpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InteractionClient struct {
	client interactionpb.InteractionServiceClient
	conn   *grpc.ClientConn
}

func NewInteractionClient(address string) (*InteractionClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to interaction service: %w", err)
	}

	client := interactionpb.NewInteractionServiceClient(conn)

	return &InteractionClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *InteractionClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// VideoInteraction represents interaction data for a video
type VideoInteraction struct {
	VideoID       uuid.UUID
	LikeCount     int64
	CommentCount  int64
	BookmarkCount int64
	ViewCount     int64
	IsLiked       bool
	IsBookmarked  bool
}

// GetVideoInteraction fetches interaction data for a single video
func (c *InteractionClient) GetVideoInteraction(ctx context.Context, videoID uuid.UUID, userID *uuid.UUID) (*VideoInteraction, error) {
	req := &interactionpb.GetVideoInteractionRequest{
		VideoId: videoID.String(),
	}

	if userID != nil {
		userIDStr := userID.String()
		req.UserId = &userIDStr
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetVideoInteraction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video interaction: %w", err)
	}

	return &VideoInteraction{
		VideoID:       videoID,
		LikeCount:     resp.LikeCount,
		CommentCount:  resp.CommentCount,
		BookmarkCount: resp.BookmarkCount,
		ViewCount:     resp.ViewCount,
		IsLiked:       resp.IsLiked,
		IsBookmarked:  resp.IsBookmarked,
	}, nil
}
