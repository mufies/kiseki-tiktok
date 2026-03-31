package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DBHost                     string
	DBPort                     string
	DBUser                     string
	DBPassword                 string
	DBName                     string
	MinioEndpoint              string
	MinioPublicEndpoint        string
	MinioPresignedEndpoint     string
	MinioAccessKey             string
	MinioSecretKey             string
	MinioBucket                string
	MinioThumbnailsBucket      string
	MinioUseSSL                bool
	MinioPresignedUseSSL       bool
	ServerPort                 string
	UserServiceGRPCAddr        string
	InteractionServiceGRPCAddr string
}

func Load() *Config {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file, using enviroment variables")
	}

	thumbnailsBucket := os.Getenv("MINIO_THUMBNAILS_BUCKET")
	if thumbnailsBucket == "" {
		thumbnailsBucket = os.Getenv("MINIO_BUCKET") + "-thumbnails"
	}

	minioPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if minioPublicEndpoint == "" {
		minioPublicEndpoint = os.Getenv("MINIO_ENDPOINT")
	}

	// Endpoint for generating presigned URLs (must be accessible from both container and browser)
	minioPresignedEndpoint := os.Getenv("MINIO_PRESIGNED_ENDPOINT")
	if minioPresignedEndpoint == "" {
		minioPresignedEndpoint = os.Getenv("MINIO_ENDPOINT")
	}

	return &Config{
		DBHost:                 os.Getenv("DB_HOST"),
		DBPort:                 os.Getenv("DB_PORT"),
		DBUser:                 os.Getenv("DB_USER"),
		DBPassword:             os.Getenv("DB_PASSWORD"),
		DBName:                 os.Getenv("DB_NAME"),
		MinioEndpoint:          os.Getenv("MINIO_ENDPOINT"),
		MinioPublicEndpoint:    minioPublicEndpoint,
		MinioPresignedEndpoint: minioPresignedEndpoint,
		MinioAccessKey:         os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:         os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:            os.Getenv("MINIO_BUCKET"),
		MinioThumbnailsBucket:  thumbnailsBucket,
		MinioUseSSL:            os.Getenv("MINIO_USE_SSL") == "true",
		MinioPresignedUseSSL:   os.Getenv("MINIO_PRESIGNED_USE_SSL") == "true",
		ServerPort:             os.Getenv("SERVER_PORT"),
	}
}

func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
