package grpc

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/kiseki/stream-service/internal/grpc/streampb"
	"github.com/kiseki/stream-service/internal/model"
	"github.com/kiseki/stream-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamGRPCServer implements the StreamService gRPC service
type StreamGRPCServer struct {
	streampb.UnimplementedStreamServiceServer
	repo repository.StreamRepository
}

// NewStreamGRPCServer creates a new gRPC server
func NewStreamGRPCServer(repo repository.StreamRepository) *StreamGRPCServer {
	return &StreamGRPCServer{
		repo: repo,
	}
}

// GetStream retrieves a stream by ID
func (s *StreamGRPCServer) GetStream(ctx context.Context, req *streampb.GetStreamRequest) (*streampb.GetStreamResponse, error) {
	log.Printf("[gRPC] GetStream called with ID: %s", req.StreamId)

	streamID, err := uuid.Parse(req.StreamId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid stream ID: %v", err)
	}

	stream, err := s.repo.FindByID(streamID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "stream not found: %v", err)
	}

	return &streampb.GetStreamResponse{
		Stream: streamModelToProto(stream),
	}, nil
}

// GetStreamByKey retrieves a stream by stream key (for RTMP authentication)
func (s *StreamGRPCServer) GetStreamByKey(ctx context.Context, req *streampb.GetStreamByKeyRequest) (*streampb.GetStreamByKeyResponse, error) {
	log.Printf("[gRPC] GetStreamByKey called with key: %s", req.StreamKey)

	stream, err := s.repo.FindByStreamKey(req.StreamKey)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "stream not found: %v", err)
	}

	return &streampb.GetStreamByKeyResponse{
		Stream: streamModelToProto(stream),
	}, nil
}

// GetLiveStreams retrieves all live streams
func (s *StreamGRPCServer) GetLiveStreams(ctx context.Context, req *streampb.GetLiveStreamsRequest) (*streampb.GetLiveStreamsResponse, error) {
	log.Printf("[gRPC] GetLiveStreams called - limit: %d, offset: %d", req.Limit, req.Offset)

	limit := int(req.Limit)
	offset := int(req.Offset)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	streams, err := s.repo.FindLiveStreams(limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get live streams: %v", err)
	}

	protoStreams := make([]*streampb.Stream, len(streams))
	for i, stream := range streams {
		protoStreams[i] = streamModelToProto(stream)
	}

	return &streampb.GetLiveStreamsResponse{
		Streams:    protoStreams,
		TotalCount: int32(len(protoStreams)),
	}, nil
}

// GetUserStreams retrieves all streams for a user
func (s *StreamGRPCServer) GetUserStreams(ctx context.Context, req *streampb.GetUserStreamsRequest) (*streampb.GetUserStreamsResponse, error) {
	log.Printf("[gRPC] GetUserStreams called for user: %s", req.UserId)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	streams, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user streams: %v", err)
	}

	protoStreams := make([]*streampb.Stream, len(streams))
	for i, stream := range streams {
		protoStreams[i] = streamModelToProto(stream)
	}

	return &streampb.GetUserStreamsResponse{
		Streams:    protoStreams,
		TotalCount: int32(len(protoStreams)),
	}, nil
}

// UpdateStreamStatus updates the status of a stream
func (s *StreamGRPCServer) UpdateStreamStatus(ctx context.Context, req *streampb.UpdateStreamStatusRequest) (*streampb.UpdateStreamStatusResponse, error) {
	log.Printf("[gRPC] UpdateStreamStatus called - ID: %s, Status: %s", req.StreamId, req.Status)

	streamID, err := uuid.Parse(req.StreamId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid stream ID: %v", err)
	}

	// Validate status
	var streamStatus model.StreamStatus
	switch req.Status {
	case "offline":
		streamStatus = model.StreamStatusOffline
	case "live":
		streamStatus = model.StreamStatusLive
	case "ending":
		streamStatus = model.StreamStatusEnding
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid status: %s", req.Status)
	}

	if err := s.repo.UpdateStatus(streamID, streamStatus); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update status: %v", err)
	}

	return &streampb.UpdateStreamStatusResponse{
		Success: true,
		Message: "Stream status updated successfully",
	}, nil
}

// GetStreamStats retrieves statistics for a stream
func (s *StreamGRPCServer) GetStreamStats(ctx context.Context, req *streampb.GetStreamStatsRequest) (*streampb.GetStreamStatsResponse, error) {
	log.Printf("[gRPC] GetStreamStats called for stream: %s", req.StreamId)

	streamID, err := uuid.Parse(req.StreamId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid stream ID: %v", err)
	}

	stream, err := s.repo.FindByID(streamID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "stream not found: %v", err)
	}

	// Calculate duration
	var durationSeconds int64
	if stream.StartedAt != nil && stream.EndedAt != nil {
		durationSeconds = stream.EndedAt.Unix() - stream.StartedAt.Unix()
	}

	stats := &streampb.StreamStats{
		StreamId:        stream.ID.String(),
		ViewerCount:     stream.ViewerCount,
		PeakViewers:     0, // TODO: Get from Redis
		TotalViewers:    0, // TODO: Get from Redis
		DurationSeconds: durationSeconds,
		AverageBitrate:  0, // TODO: Calculate
		VideoPackets:    0, // TODO: Get from RTMP handler
		AudioPackets:    0, // TODO: Get from RTMP handler
	}

	return &streampb.GetStreamStatsResponse{
		Stats: stats,
	}, nil
}

// Helper function to convert model.Stream to streampb.Stream
func streamModelToProto(stream *model.Stream) *streampb.Stream {
	protoStream := &streampb.Stream{
		Id:           stream.ID.String(),
		UserId:       stream.UserID.String(),
		StreamKey:    stream.StreamKey,
		Title:        stream.Title,
		Description:  stream.Description,
		ThumbnailUrl: stream.ThumbnailURL,
		Status:       string(stream.Status),
		ViewerCount:  stream.ViewerCount,
		SaveVod:      stream.SaveVOD,
		CreatedAt:    stream.CreatedAt.Unix(),
		UpdatedAt:    stream.UpdatedAt.Unix(),
	}

	if stream.StartedAt != nil {
		protoStream.StartedAt = stream.StartedAt.Unix()
	}
	if stream.EndedAt != nil {
		protoStream.EndedAt = stream.EndedAt.Unix()
	}

	return protoStream
}
