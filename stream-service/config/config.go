package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// MinIO/S3 Storage for VOD and recordings
	MinioEndpoint          string
	MinioPublicEndpoint    string
	MinioPresignedEndpoint string
	MinioAccessKey         string
	MinioSecretKey         string
	MinioStreamsBucket     string // HLS segments and VOD files
	MinioThumbnailsBucket  string
	MinioUseSSL            bool
	MinioPresignedUseSSL   bool

	// RTMP Server Configuration
	RTMPPort          string
	RTMPChunkSize     int
	RTMPMaxConns      int
	HLSSegmentTime    int // seconds per HLS segment
	HLSPlaylistLength int // number of segments in playlist

	// Kafka Configuration
	KafkaBrokers []string
	KafkaTopics  KafkaTopics

	// Redis Configuration (for real-time viewer counts, chat)
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// HTTP Server
	ServerPort string
	GRPCPort   string

	// External Service Addresses
	UserServiceGRPCAddr        string
	VideoServiceGRPCAddr       string
	NotificationServiceGRPCAddr string

	// Transcoding Configuration
	EnableTranscoding bool
	FFmpegPath        string
	TranscodingPresets []TranscodingPreset
}

type KafkaTopics struct {
	StreamStarted    string
	StreamEnded      string
	ViewerJoined     string
	ViewerLeft       string
	StreamUpdate     string
	VODReady         string
}

type TranscodingPreset struct {
	Name       string
	Resolution string // e.g., "1920x1080", "1280x720"
	VideoBitrate string // e.g., "5000k", "2500k"
	AudioBitrate string // e.g., "128k", "96k"
}

func Load() *Config {
	// Try to load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set defaults
	thumbnailsBucket := getEnv("MINIO_THUMBNAILS_BUCKET", getEnv("MINIO_STREAMS_BUCKET", "streams")+"-thumbnails")
	minioPublicEndpoint := getEnv("MINIO_PUBLIC_ENDPOINT", getEnv("MINIO_ENDPOINT", "localhost:9000"))
	minioPresignedEndpoint := getEnv("MINIO_PRESIGNED_ENDPOINT", getEnv("MINIO_ENDPOINT", "localhost:9000"))

	// Kafka brokers
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	// Default transcoding presets
	transcodingPresets := []TranscodingPreset{
		{
			Name:         "1080p",
			Resolution:   "1920x1080",
			VideoBitrate: "5000k",
			AudioBitrate: "128k",
		},
		{
			Name:         "720p",
			Resolution:   "1280x720",
			VideoBitrate: "2500k",
			AudioBitrate: "128k",
		},
		{
			Name:         "480p",
			Resolution:   "854x480",
			VideoBitrate: "1000k",
			AudioBitrate: "96k",
		},
		{
			Name:         "360p",
			Resolution:   "640x360",
			VideoBitrate: "600k",
			AudioBitrate: "64k",
		},
	}

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "stream_service"),

		// MinIO
		MinioEndpoint:          getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioPublicEndpoint:    minioPublicEndpoint,
		MinioPresignedEndpoint: minioPresignedEndpoint,
		MinioAccessKey:         getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:         getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioStreamsBucket:     getEnv("MINIO_STREAMS_BUCKET", "streams"),
		MinioThumbnailsBucket:  thumbnailsBucket,
		MinioUseSSL:            getEnv("MINIO_USE_SSL", "false") == "true",
		MinioPresignedUseSSL:   getEnv("MINIO_PRESIGNED_USE_SSL", "false") == "true",

		// RTMP
		RTMPPort:          getEnv("RTMP_PORT", "1935"),
		RTMPChunkSize:     getEnvAsInt("RTMP_CHUNK_SIZE", 4096),
		RTMPMaxConns:      getEnvAsInt("RTMP_MAX_CONNS", 1000),
		HLSSegmentTime:    getEnvAsInt("HLS_SEGMENT_TIME", 6),
		HLSPlaylistLength: getEnvAsInt("HLS_PLAYLIST_LENGTH", 5),

		// Kafka
		KafkaBrokers: kafkaBrokers,
		KafkaTopics: KafkaTopics{
			StreamStarted: getEnv("KAFKA_TOPIC_STREAM_STARTED", "stream.started"),
			StreamEnded:   getEnv("KAFKA_TOPIC_STREAM_ENDED", "stream.ended"),
			ViewerJoined:  getEnv("KAFKA_TOPIC_VIEWER_JOINED", "stream.viewer.joined"),
			ViewerLeft:    getEnv("KAFKA_TOPIC_VIEWER_LEFT", "stream.viewer.left"),
			StreamUpdate:  getEnv("KAFKA_TOPIC_STREAM_UPDATE", "stream.updated"),
			VODReady:      getEnv("KAFKA_TOPIC_VOD_READY", "stream.vod.ready"),
		},

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// HTTP/gRPC Servers
		ServerPort: getEnv("SERVER_PORT", "8083"),
		GRPCPort:   getEnv("GRPC_PORT", "50055"),

		// External Services
		UserServiceGRPCAddr:        getEnv("USER_SERVICE_GRPC_ADDR", "localhost:50053"),
		VideoServiceGRPCAddr:       getEnv("VIDEO_SERVICE_GRPC_ADDR", "localhost:50052"),
		NotificationServiceGRPCAddr: getEnv("NOTIFICATION_SERVICE_GRPC_ADDR", "localhost:50051"),

		// Transcoding
		EnableTranscoding:  getEnv("ENABLE_TRANSCODING", "true") == "true",
		FFmpegPath:         getEnv("FFMPEG_PATH", "ffmpeg"),
		TranscodingPresets: transcodingPresets,
	}
}

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Successfully connected to database: %s", cfg.DBName)
	return db
}

// Helper functions
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	return valStr == "true" || valStr == "1"
}
