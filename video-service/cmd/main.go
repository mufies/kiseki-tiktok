package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kiseki/video-service/config"
	grpcserver "github.com/kiseki/video-service/internal/grpc"
	"github.com/kiseki/video-service/internal/grpc/interactionclient"
	"github.com/kiseki/video-service/internal/grpc/userpb"
	"github.com/kiseki/video-service/internal/grpc/videopb"
	"github.com/kiseki/video-service/internal/handler"
	"github.com/kiseki/video-service/internal/model"
	"github.com/kiseki/video-service/internal/repository"
	"github.com/kiseki/video-service/internal/service"
	"github.com/kiseki/video-service/internal/authorization"
	"github.com/kiseki/video-service/internal/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	cfg := config.Load()

	db := config.ConnectDB(cfg)
	db.AutoMigrate(&model.Video{})

	// MinIO client for operations (upload, download, etc) - uses internal endpoint
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatal("Failed to connect Minio:", err)
	}

	exist, _ := minioClient.BucketExists(context.Background(), cfg.MinioBucket)
	if !exist {
		minioClient.MakeBucket(context.Background(), cfg.MinioBucket, minio.MakeBucketOptions{})
		log.Println("Bucket created:", cfg.MinioBucket)
	}

	thumbnailsExist, _ := minioClient.BucketExists(context.Background(), cfg.MinioThumbnailsBucket)
	if !thumbnailsExist {
		minioClient.MakeBucket(context.Background(), cfg.MinioThumbnailsBucket, minio.MakeBucketOptions{})
		log.Println("Thumbnails bucket created:", cfg.MinioThumbnailsBucket)
	}

	// MinIO client for generating presigned URLs - uses endpoint accessible from browser
	log.Printf("Creating presigned client with endpoint: %s", cfg.MinioPresignedEndpoint)
	minioPresignedClient, err := minio.New(cfg.MinioPresignedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioPresignedUseSSL,
	})
	if err != nil {
		log.Printf("Warning: Failed to create presigned client, will use main client: %v", err)
		minioPresignedClient = minioClient
	} else {
		log.Printf("Successfully created presigned client with endpoint: %s", cfg.MinioPresignedEndpoint)
	}

	repo := repository.NewVideoRepository(db)
	storageClient := storage.NewMinioStorageClient(minioClient)
	storageClient.SetPresignedClient(minioPresignedClient)
	authService := authorization.NewVideoAuthorizationService()

	// Connect to User Service for gRPC
	// Use service name for Docker, fallback to localhost for local development
	userServiceAddr := getEnv("USER_SERVICE_GRPC_ADDR", "user-service:50053")
	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to user service at %s: %v", userServiceAddr, err)
	}
	var userClient userpb.UserServiceClient
	if userConn != nil {
		userClient = userpb.NewUserServiceClient(userConn)
		log.Printf("Connected to user service at %s", userServiceAddr)
	}

	// Interaction service HTTP endpoint
	// Use service name for Docker, fallback to localhost for local development
	interactionServiceURL := getEnv("INTERACTION_SERVICE_URL", "http://interaction-service:8084")
	log.Printf("Using interaction service at %s", interactionServiceURL)

	svc := service.NewVideoService(
		repo,
		storageClient,
		authService,
		cfg.MinioBucket,
		cfg.MinioThumbnailsBucket,
		cfg.MinioPresignedEndpoint,
		cfg.MinioPublicEndpoint,
		userClient,
		interactionServiceURL,
	)
	h := handler.NewVideoHandler(svc)

	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("gRPC listen error: %v", err)
		}
		gs := grpc.NewServer()
		videopb.RegisterVideoServiceServer(gs, grpcserver.NewVideoGRPCServer(repo, userClient))
		log.Println("Video gRPC server listening on :50052")
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	r := gin.Default()
	r.MaxMultipartMemory = 500 << 20

	v1 := r.Group("/videos")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/upload", h.Upload)
		v1.GET("/:id/presigned-url", h.GetPresignedURL)
		v1.PATCH("/:id", h.UpdateVideo)
		v1.GET("/:id", h.GetByID)
		v1.GET("/user/:userId", h.GetByOwner)
		v1.DELETE("/:id", h.Delete)
	}

	log.Println("Video HTTP server running on :" + cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}
