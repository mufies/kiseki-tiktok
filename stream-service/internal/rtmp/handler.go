package rtmp

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	rtmp "github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"

	"github.com/google/uuid"
	"github.com/kiseki/stream-service/internal/service"
	"github.com/kiseki/stream-service/internal/transcoder"
)

// StreamHandler handles RTMP stream events and integrates with StreamService
// This is a shared handler that manages all active streams
type StreamHandler struct {
	streamService     *service.StreamService
	transcoderManager *transcoder.Manager
	ctx               context.Context

	// Track active streams globally
	mu            sync.RWMutex
	activeStreams map[string]*ActiveStream // key: stream_key, value: stream info
}

// ConnectionHandler handles events for a single RTMP connection
// Each connection gets its own instance to properly track per-stream statistics
type ConnectionHandler struct {
	rtmp.DefaultHandler
	sharedHandler *StreamHandler

	// Track this connection's stream
	mu           sync.RWMutex
	streamKey    string
	streamID     uuid.UUID
	flvWriter    *transcoder.FLVWriter
	lastVideoTs  uint32
	lastAudioTs  uint32
}

// ActiveStream holds information about an active RTMP stream
type ActiveStream struct {
	StreamID      uuid.UUID
	StreamKey     string
	UserID        uuid.UUID
	StartTime     time.Time
	VideoPackets  int64
	AudioPackets  int64
	TotalBytes    int64
	LastPacketAt  time.Time
}

func NewStreamHandler(streamService *service.StreamService, transcoderMgr *transcoder.Manager) *StreamHandler {
	return &StreamHandler{
		streamService:     streamService,
		transcoderManager: transcoderMgr,
		ctx:               context.Background(),
		activeStreams:     make(map[string]*ActiveStream),
	}
}

// NewConnectionHandler creates a new handler for a single connection
func (h *StreamHandler) NewConnectionHandler() *ConnectionHandler {
	return &ConnectionHandler{
		sharedHandler: h,
	}
}

// OnServe is called when a new connection is established
func (h *ConnectionHandler) OnServe(conn *rtmp.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RTMP] PANIC in OnServe: %v", r)
		}
	}()
	log.Printf("[RTMP] New connection established")
}

// OnConnect is called when client sends connect command
func (h *ConnectionHandler) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RTMP] PANIC in OnConnect: %v", r)
		}
	}()
	log.Printf("[RTMP] ✅ OnConnect SUCCESS - App: %s, FlashVer: %s", cmd.Command.App, cmd.Command.FlashVer)
	return nil
}

// OnCreateStream is called when client creates a stream
func (h *ConnectionHandler) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	log.Printf("[RTMP] OnCreateStream")
	return nil
}

// OnReleaseStream is called to release a stream (Flash compatibility)
func (h *ConnectionHandler) OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error {
	log.Printf("[RTMP] OnReleaseStream: %s", cmd.StreamName)
	return nil
}

// OnDeleteStream is called when stream is being deleted
func (h *ConnectionHandler) OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error {
	log.Printf("[RTMP] OnDeleteStream: stream ID %d", cmd.StreamID)
	return nil
}

// OnPublish is called when client starts publishing a stream
func (h *ConnectionHandler) OnPublish(streamCtx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	streamKey := cmd.PublishingName
	log.Printf("[RTMP] 📡 OnPublish - Stream Key: %s, Type: %s", streamKey, cmd.PublishingType)

	sh := h.sharedHandler

	// Validate stream key with service
	stream, err := sh.streamService.GetStreamByKey(sh.ctx, streamKey)
	if err != nil {
		log.Printf("[RTMP] ❌ Invalid stream key: %s - %v", streamKey, err)
		return fmt.Errorf("invalid stream key: %w", err)
	}

	log.Printf("[RTMP] ✅ Stream key validated - Stream ID: %s, User ID: %s, Title: %s",
		stream.ID, stream.UserID, stream.Title)

	// Check if stream can go live
	if !stream.CanGoLive() {
		log.Printf("[RTMP] ❌ Stream cannot go live - Current status: %s", stream.Status)
		return fmt.Errorf("stream cannot go live: status is %s", stream.Status)
	}

	// Start the stream in the service
	if err := sh.streamService.StartStream(sh.ctx, stream.ID); err != nil {
		log.Printf("[RTMP] ❌ Failed to start stream: %v", err)
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Track active stream in shared handler
	sh.mu.Lock()
	sh.activeStreams[streamKey] = &ActiveStream{
		StreamID:     stream.ID,
		StreamKey:    streamKey,
		UserID:       stream.UserID,
		StartTime:    time.Now(),
		LastPacketAt: time.Now(),
	}
	sh.mu.Unlock()

	// Start HLS transcoder
	hlsTranscoder, err := sh.transcoderManager.StartTranscoder(stream.ID, streamKey)
	if err != nil {
		log.Printf("[RTMP] ⚠️  Failed to start transcoder: %v (stream will continue without HLS)", err)
		// Don't fail the stream if transcoder fails - continue without HLS
	} else {
		// Create FLV writer to pipe data to FFmpeg stdin
		h.mu.Lock()
		h.flvWriter = hlsTranscoder.GetFLVWriter()
		h.streamID = stream.ID
		h.mu.Unlock()

		log.Printf("[RTMP] 📹 HLS transcoder started - Playlist: %s", hlsTranscoder.GetPlaylistURL())
	}

	// Store stream key for this connection
	h.mu.Lock()
	h.streamKey = streamKey
	h.mu.Unlock()

	log.Printf("[RTMP] 🎬 Stream started successfully - ID: %s, Key: %s", stream.ID, streamKey)
	return nil
}

// OnAudio is called when audio data is received
func (h *ConnectionHandler) OnAudio(timestamp uint32, payload io.Reader) error {
	// Read audio data
	data, err := io.ReadAll(payload)
	if err != nil {
		return err
	}

	// Get this connection's stream key and FLV writer
	h.mu.RLock()
	streamKey := h.streamKey
	flvWriter := h.flvWriter
	h.mu.RUnlock()

	if streamKey == "" {
		// No stream published yet, ignore
		return nil
	}

	// Update statistics for THIS stream only
	sh := h.sharedHandler
	sh.mu.Lock()
	stream, ok := sh.activeStreams[streamKey]
	if ok {
		stream.AudioPackets++
		stream.TotalBytes += int64(len(data))
		stream.LastPacketAt = time.Now()

		// Log statistics every 100 audio packets
		if stream.AudioPackets%100 == 0 {
			duration := time.Since(stream.StartTime)
			bitrate := float64(stream.TotalBytes*8) / duration.Seconds() / 1000000 // Mbps
			log.Printf("[RTMP] 🎵 Stream %s - Audio: %d pkts, Video: %d pkts, Duration: %s, Bitrate: %.2f Mbps",
				stream.StreamKey,
				stream.AudioPackets,
				stream.VideoPackets,
				duration.Round(time.Second),
				bitrate,
			)
		}
	}
	sh.mu.Unlock()

	// Forward audio data to transcoder for HLS conversion
	if flvWriter != nil {
		h.mu.Lock()
		h.lastAudioTs = timestamp
		h.mu.Unlock()

		if err := flvWriter.WriteAudioTag(timestamp, data); err != nil {
			log.Printf("[RTMP] ⚠️  Failed to write audio to transcoder: %v", err)
			// Don't return error - continue streaming even if transcoding fails
		}
	}

	return nil
}

// OnVideo is called when video data is received
func (h *ConnectionHandler) OnVideo(timestamp uint32, payload io.Reader) error {
	// Read video data
	data, err := io.ReadAll(payload)
	if err != nil {
		return err
	}

	// Get this connection's stream key and FLV writer
	h.mu.RLock()
	streamKey := h.streamKey
	flvWriter := h.flvWriter
	h.mu.RUnlock()

	if streamKey == "" {
		// No stream published yet, ignore
		return nil
	}

	// Update statistics for THIS stream only
	sh := h.sharedHandler
	sh.mu.Lock()
	stream, ok := sh.activeStreams[streamKey]
	if ok {
		stream.VideoPackets++
		stream.TotalBytes += int64(len(data))
		stream.LastPacketAt = time.Now()

		// Log statistics every 30 video packets (roughly 1 second at 30fps)
		if stream.VideoPackets%30 == 0 {
			duration := time.Since(stream.StartTime)
			bitrate := float64(stream.TotalBytes*8) / duration.Seconds() / 1000000 // Mbps
			log.Printf("[RTMP] 📹 Stream %s - Video: %d pkts, Audio: %d pkts, Duration: %s, Bitrate: %.2f Mbps",
				stream.StreamKey,
				stream.VideoPackets,
				stream.AudioPackets,
				duration.Round(time.Second),
				bitrate,
			)
		}
	}
	sh.mu.Unlock()

	// Forward video data to transcoder for HLS conversion
	if flvWriter != nil {
		h.mu.Lock()
		h.lastVideoTs = timestamp
		h.mu.Unlock()

		if err := flvWriter.WriteVideoTag(timestamp, data); err != nil {
			log.Printf("[RTMP] ⚠️  Failed to write video to transcoder: %v", err)
			// Don't return error - continue streaming even if transcoding fails
		}
	}

	return nil
}

// OnPlay is called when client starts playing a stream (not used for publishing)
func (h *ConnectionHandler) OnPlay(streamCtx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPlay) error {
	log.Printf("[RTMP] OnPlay: %s", cmd.StreamName)
	// Not implemented - we only support publishing, not playback
	return fmt.Errorf("playback not supported")
}

// OnFCPublish is called for Flash Media Server compatibility (before publish)
func (h *ConnectionHandler) OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error {
	log.Printf("[RTMP] OnFCPublish")
	return nil
}

// OnFCUnpublish is called for Flash Media Server compatibility (after unpublish)
func (h *ConnectionHandler) OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error {
	log.Printf("[RTMP] OnFCUnpublish")
	return nil
}

// OnSetDataFrame is called when metadata is received
func (h *ConnectionHandler) OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error {
	log.Printf("[RTMP] OnSetDataFrame: %+v", data)
	// TODO: Store stream metadata (resolution, bitrate, codec, etc.)
	return nil
}

// OnUnknownMessage is called when an unknown message type is received
func (h *ConnectionHandler) OnUnknownMessage(timestamp uint32, msg message.Message) error {
	log.Printf("[RTMP] OnUnknownMessage: %+v", msg)
	return nil
}

// OnUnknownCommandMessage is called when an unknown command message is received
func (h *ConnectionHandler) OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error {
	log.Printf("[RTMP] OnUnknownCommandMessage: %s", cmd.CommandName)
	return nil
}

// OnUnknownDataMessage is called when an unknown data message is received
func (h *ConnectionHandler) OnUnknownDataMessage(timestamp uint32, data *message.DataMessage) error {
	log.Printf("[RTMP] OnUnknownDataMessage: %+v", data)
	return nil
}

// OnClose is called when the stream connection is closed
func (h *ConnectionHandler) OnClose() {
	log.Printf("[RTMP] 🔌 Connection closing")

	// Get this connection's stream key
	h.mu.RLock()
	streamKey := h.streamKey
	h.mu.RUnlock()

	if streamKey == "" {
		// No stream was published on this connection
		log.Printf("[RTMP] Connection closed without publishing a stream")
		return
	}

	sh := h.sharedHandler

	// Get and remove the stream from active streams
	sh.mu.Lock()
	stream, ok := sh.activeStreams[streamKey]
	if ok {
		delete(sh.activeStreams, streamKey)
	}
	sh.mu.Unlock()

	if !ok {
		log.Printf("[RTMP] Stream key %s not found in active streams", streamKey)
		return
	}

	// Calculate and log statistics
	duration := time.Since(stream.StartTime)
	bitrate := float64(stream.TotalBytes*8) / duration.Seconds() / 1000000 // Mbps

	log.Printf("[RTMP] 🛑 Ending stream - ID: %s, Key: %s", stream.StreamID, stream.StreamKey)
	log.Printf("[RTMP]    Duration: %s", duration.Round(time.Second))
	log.Printf("[RTMP]    Video Packets: %d", stream.VideoPackets)
	log.Printf("[RTMP]    Audio Packets: %d", stream.AudioPackets)
	log.Printf("[RTMP]    Total Data: %.2f MB", float64(stream.TotalBytes)/(1024*1024))
	log.Printf("[RTMP]    Avg Bitrate: %.2f Mbps", bitrate)

	// Stop transcoder if it was running
	if sh.transcoderManager != nil {
		if err := sh.transcoderManager.StopTranscoder(stream.StreamID); err != nil {
			log.Printf("[RTMP] ⚠️  Failed to stop transcoder: %v", err)
		}
	}

	// End stream in service
	if err := sh.streamService.EndStream(sh.ctx, stream.StreamID); err != nil {
		log.Printf("[RTMP] ❌ Failed to end stream %s: %v", stream.StreamID, err)
	} else {
		log.Printf("[RTMP] ✅ Stream ended successfully - ID: %s", stream.StreamID)
	}

	// Clear stream key and FLV writer for this connection
	h.mu.Lock()
	h.streamKey = ""
	h.flvWriter = nil
	h.streamID = uuid.Nil
	h.mu.Unlock()
}

// GetActiveStreams returns the list of currently active streams
func (h *StreamHandler) GetActiveStreams() []*ActiveStream {
	h.mu.RLock()
	defer h.mu.RUnlock()

	streams := make([]*ActiveStream, 0, len(h.activeStreams))
	for _, stream := range h.activeStreams {
		streams = append(streams, stream)
	}
	return streams
}

// GetStreamStats returns statistics for a specific stream key
func (h *StreamHandler) GetStreamStats(streamKey string) *ActiveStream {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if stream, ok := h.activeStreams[streamKey]; ok {
		return stream
	}
	return nil
}
