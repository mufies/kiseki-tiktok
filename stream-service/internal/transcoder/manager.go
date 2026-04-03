package transcoder

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/kiseki/stream-service/config"
)

// Manager manages multiple HLS transcoders (one per active stream)
type Manager struct {
	config      *config.Config
	transcoders map[uuid.UUID]*HLSTranscoder // key: stream_id
	mu          sync.RWMutex
}

// NewManager creates a new transcoder manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:      cfg,
		transcoders: make(map[uuid.UUID]*HLSTranscoder),
	}
}

// StartTranscoder creates and starts a new transcoder for a stream
func (m *Manager) StartTranscoder(streamID uuid.UUID, streamKey string) (*HLSTranscoder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if transcoder already exists
	if _, exists := m.transcoders[streamID]; exists {
		return nil, fmt.Errorf("transcoder already exists for stream %s", streamID)
	}

	// Create new transcoder
	transcoder, err := NewHLSTranscoder(streamID, streamKey, m.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create transcoder: %w", err)
	}

	// Start transcoding
	if err := transcoder.Start(); err != nil {
		return nil, fmt.Errorf("failed to start transcoder: %w", err)
	}

	// Store transcoder
	m.transcoders[streamID] = transcoder

	log.Printf("[TranscoderManager] Started transcoder for stream %s", streamID)
	return transcoder, nil
}

// StopTranscoder stops and removes a transcoder
func (m *Manager) StopTranscoder(streamID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transcoder, exists := m.transcoders[streamID]
	if !exists {
		return fmt.Errorf("transcoder not found for stream %s", streamID)
	}

	// Stop transcoder
	if err := transcoder.Stop(); err != nil {
		log.Printf("[TranscoderManager] Error stopping transcoder for stream %s: %v", streamID, err)
	}

	// Remove from map
	delete(m.transcoders, streamID)

	log.Printf("[TranscoderManager] Stopped transcoder for stream %s", streamID)
	return nil
}

// GetTranscoder returns the transcoder for a stream
func (m *Manager) GetTranscoder(streamID uuid.UUID) (*HLSTranscoder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transcoder, exists := m.transcoders[streamID]
	if !exists {
		return nil, fmt.Errorf("transcoder not found for stream %s", streamID)
	}

	return transcoder, nil
}

// GetAllTranscoders returns all active transcoders
func (m *Manager) GetAllTranscoders() map[uuid.UUID]*HLSTranscoder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	transcoders := make(map[uuid.UUID]*HLSTranscoder, len(m.transcoders))
	for id, transcoder := range m.transcoders {
		transcoders[id] = transcoder
	}

	return transcoders
}

// StopAll stops all transcoders (used during shutdown)
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[TranscoderManager] Stopping all transcoders (%d active)", len(m.transcoders))

	for streamID, transcoder := range m.transcoders {
		if err := transcoder.Stop(); err != nil {
			log.Printf("[TranscoderManager] Error stopping transcoder for stream %s: %v", streamID, err)
		}
	}

	// Clear map
	m.transcoders = make(map[uuid.UUID]*HLSTranscoder)

	log.Printf("[TranscoderManager] All transcoders stopped")
}

// GetStats returns statistics for all transcoders
func (m *Manager) GetStats() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]map[string]interface{}, 0, len(m.transcoders))
	for _, transcoder := range m.transcoders {
		stats = append(stats, transcoder.GetStats())
	}

	return stats
}
