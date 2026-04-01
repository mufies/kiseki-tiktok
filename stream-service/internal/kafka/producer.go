package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers not configured")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Compression:  kafka.Snappy,
	}

	return &KafkaProducer{
		writer: writer,
	}, nil
}

func (p *KafkaProducer) PublishEvent(ctx context.Context, topic string, key string, value interface{}) error {
	if p.writer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic:     topic,
		Key:       []byte(key),
		Value:     valueBytes,
		Time:      time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	log.Printf("Published event to topic %s: %s", topic, key)
	return nil
}

func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// Event structures for stream events

type StreamStartedEvent struct {
	StreamID    uuid.UUID `json:"stream_id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	StartedAt   time.Time `json:"started_at"`
}

type StreamEndedEvent struct {
	StreamID     uuid.UUID `json:"stream_id"`
	UserID       uuid.UUID `json:"user_id"`
	EndedAt      time.Time `json:"ended_at"`
	Duration     int64     `json:"duration_seconds"`
	PeakViewers  int64     `json:"peak_viewers"`
	TotalViewers int64     `json:"total_viewers"`
}

type ViewerJoinedEvent struct {
	StreamID     uuid.UUID `json:"stream_id"`
	UserID       uuid.UUID `json:"user_id,omitempty"`
	ViewerCount  int64     `json:"viewer_count"`
	JoinedAt     time.Time `json:"joined_at"`
}

type ViewerLeftEvent struct {
	StreamID     uuid.UUID `json:"stream_id"`
	UserID       uuid.UUID `json:"user_id,omitempty"`
	ViewerCount  int64     `json:"viewer_count"`
	LeftAt       time.Time `json:"left_at"`
}

type StreamUpdatedEvent struct {
	StreamID    uuid.UUID `json:"stream_id"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type VODReadyEvent struct {
	StreamID     uuid.UUID `json:"stream_id"`
	VideoID      uuid.UUID `json:"video_id"`
	UserID       uuid.UUID `json:"user_id"`
	VODUrl       string    `json:"vod_url"`
	Duration     int64     `json:"duration_seconds"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
}
