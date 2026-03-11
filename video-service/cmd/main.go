package main

import (
	"context"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/kiseki/video-service/config"
	grpcserver "github.com/kiseki/video-service/internal/grpc"
	"github.com/kiseki/video-service/internal/grpc/videopb"
	"github.com/kiseki/video-service/internal/handler"
	"github.com/kiseki/video-service/internal/model"
	"github.com/kiseki/video-service/internal/repository"
	"github.com/kiseki/video-service/internal/service"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	db := config.ConnectDB(cfg)
	db.AutoMigrate(&model.Video{})

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

	repo := repository.NewVideoRepository(db)
	svc := service.NewVideoService(repo, minioClient, cfg.MinioBucket)
	h := handler.NewVideoHandler(svc)

	// ── gRPC Server (port 9091) ───────────────────────────────────────────────
	go func() {
		lis, err := net.Listen("tcp", ":9091")
		if err != nil {
			log.Fatalf("gRPC listen error: %v", err)
		}
		gs := grpc.NewServer()
		videopb.RegisterVideoServiceServer(gs, grpcserver.NewVideoGRPCServer(repo))
		log.Println("Video gRPC server listening on :9091")
		if err := gs.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	// ── HTTP REST Server ──────────────────────────────────────────────────────
	r := gin.Default()
	r.MaxMultipartMemory = 500 << 20

	v1 := r.Group("/videos")
	{
		v1.POST("/upload", h.Upload)
		v1.GET("/:id", h.GetByID)
		v1.GET("/user/:userId", h.GetByOwner)
	}

	log.Println("Video HTTP server running on :" + cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}
