package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kiseki/stream-service/config"
	grpcserver "github.com/kiseki/stream-service/internal/grpc"
	"github.com/kiseki/stream-service/internal/grpc/streampb"
	"github.com/kiseki/stream-service/internal/handler"
	"github.com/kiseki/stream-service/internal/kafka"
	"github.com/kiseki/stream-service/internal/model"
	"github.com/kiseki/stream-service/internal/repository"
	"github.com/kiseki/stream-service/internal/rtmp"
	"github.com/kiseki/stream-service/internal/service"
	"github.com/kiseki/stream-service/internal/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load configuration
	cfg := config.Load()
	log.Println("Configuration loaded successfully")

	// Connect to database
	db := config.ConnectDB(cfg)

	// Auto migrate database schema
	if err := db.AutoMigrate(&model.Stream{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// Initialize MinIO client for storage operations
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}
	log.Println("Connected to MinIO successfully")

	// Create buckets if they don't exist
	ctx := context.Background()
	buckets := []string{cfg.MinioStreamsBucket, cfg.MinioThumbnailsBucket}
	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			log.Printf("Warning: Failed to check bucket %s: %v", bucket, err)
			continue
		}
		if !exists {
			if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				log.Printf("Warning: Failed to create bucket %s: %v", bucket, err)
			} else {
				log.Printf("Bucket created: %s", bucket)
			}
		}
	}

	// MinIO client for presigned URLs
	minioPresignedClient, err := minio.New(cfg.MinioPresignedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioPresignedUseSSL,
	})
	if err != nil {
		log.Printf("Warning: Failed to create presigned client, using main client: %v", err)
		minioPresignedClient = minioClient
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis successfully")
	}

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
	} else {
		log.Println("Kafka producer initialized successfully")
		defer kafkaProducer.Close()
	}

	// Initialize repository
	repo := repository.NewStreamRepository(db)

	// Initialize storage client
	storageClient := storage.NewMinioStorageClient(minioClient)
	storageClient.SetPresignedClient(minioPresignedClient)

	// Initialize stream service
	streamService := service.NewStreamService(
		repo,
		storageClient,
		redisClient,
		kafkaProducer,
		cfg,
	)

	// Initialize HTTP handler
	streamHandler := handler.NewStreamHandler(streamService)

	// Start gRPC server in background
	go startGRPCServer(cfg, repo)

	// Initialize and start RTMP server for stream ingestion
	rtmpHandler := rtmp.NewStreamHandler(streamService)
	rtmpServer := rtmp.NewServer(":"+cfg.RTMPPort, rtmpHandler)
	go func() {
		log.Printf("Starting RTMP server on port %s", cfg.RTMPPort)
		if err := rtmpServer.Start(); err != nil {
			log.Printf("RTMP server error: %v", err)
		}
	}()
	log.Println("RTMP server started")


	// Setup HTTP server with Gin
	r := gin.Default()
	r.Use(corsMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "stream-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	v1 := r.Group("/streams")
	{
		// Stream management
		v1.POST("", streamHandler.CreateStream)       // Create new stream
		v1.GET("/:id", streamHandler.GetStream)       // Get stream details
		v1.PATCH("/:id", streamHandler.UpdateStream)  // Update stream info
		v1.DELETE("/:id", streamHandler.DeleteStream) // Delete stream

		// Stream lifecycle
		v1.POST("/:id/start", streamHandler.StartStream) // Start streaming
		v1.POST("/:id/end", streamHandler.EndStream)     // End streaming

		// Stream discovery
		v1.GET("/live", streamHandler.GetLiveStreams)         // List all live streams
		v1.GET("/user/:userId", streamHandler.GetUserStreams) // Get user's streams

		// Stream playback
		v1.GET("/:id/playback", streamHandler.GetPlaybackURL) // Get HLS playback URL

		// Viewer management
		v1.POST("/:id/viewers/join", streamHandler.JoinStream)   // Join stream (increment viewers)
		v1.POST("/:id/viewers/leave", streamHandler.LeaveStream) // Leave stream (decrement viewers)
	}

	// Start HTTP server
	log.Printf("Stream HTTP server starting on :%s", cfg.ServerPort)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown RTMP server
	if err := rtmpServer.Stop(); err != nil {
		log.Printf("RTMP server shutdown error: %v", err)
	}

	log.Println("Server exited")
}

func startGRPCServer(cfg *config.Config, repo repository.StreamRepository) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register gRPC service
	streampb.RegisterStreamServiceServer(grpcServer, grpcserver.NewStreamGRPCServer(repo))

	log.Printf("Stream gRPC server listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
