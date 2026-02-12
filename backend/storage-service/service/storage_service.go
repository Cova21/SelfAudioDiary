package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// StorageService handles file storage operations
type StorageService struct {
	minioClient *minio.Client
	bucketName  string
}

// NewStorageService creates a new storage service
func NewStorageService(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*StorageService, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	logger.Info("Connected to MinIO", zap.String("endpoint", endpoint))

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created MinIO bucket", zap.String("bucket", bucketName))
	}

	return &StorageService{
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

// GenerateUploadURL generates a presigned URL for uploading a file
func (s *StorageService) GenerateUploadURL(userID, filename, contentType string, fileSize int64) (string, string, error) {
	fileID := uuid.New().String()
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	expiry := 15 * time.Minute

	url, err := s.minioClient.PresignedPutObject(ctx, s.bucketName, objectName, expiry)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	logger.Info("Generated upload URL",
		zap.String("file_id", fileID),
		zap.String("user_id", userID),
	)

	return url.String(), fileID, nil
}

// GenerateDownloadURL generates a presigned URL for downloading a file
func (s *StorageService) GenerateDownloadURL(fileID, userID string) (string, error) {
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	expiry := 1 * time.Hour

	url, err := s.minioClient.PresignedGetObject(ctx, s.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	logger.Info("Generated download URL",
		zap.String("file_id", fileID),
		zap.String("user_id", userID),
	)

	return url.String(), nil
}

// UploadFile uploads a file directly (for internal services)
func (s *StorageService) UploadFile(userID, filename string, content []byte, contentType string) (string, string, int64, error) {
	fileID := uuid.New().String()
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	reader := bytes.NewReader(content)

	_, err := s.minioClient.PutObject(
		ctx,
		s.bucketName,
		objectName,
		reader,
		int64(len(content)),
		minio.PutObjectOptions{ContentType: contentType},
	)

	if err != nil {
		return "", "", 0, fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("Uploaded file",
		zap.String("file_id", fileID),
		zap.String("user_id", userID),
		zap.Int("size", len(content)),
	)

	// Generate download URL
	url, err := s.GenerateDownloadURL(fileID, userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return fileID, url, int64(len(content)), nil
}

// DownloadFile downloads a file directly (for internal services)
func (s *StorageService) DownloadFile(fileID, userID string) ([]byte, string, string, int64, error) {
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	object, err := s.minioClient.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Get object info
	stat, err := object.Stat()
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to stat object: %w", err)
	}

	// Read content
	content, err := io.ReadAll(object)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to read object: %w", err)
	}

	logger.Info("Downloaded file",
		zap.String("file_id", fileID),
		zap.String("user_id", userID),
		zap.Int64("size", stat.Size),
	)

	return content, fileID, stat.ContentType, stat.Size, nil
}

// DeleteFile deletes a file
func (s *StorageService) DeleteFile(fileID, userID string) error {
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	err := s.minioClient.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("Deleted file",
		zap.String("file_id", fileID),
		zap.String("user_id", userID),
	)

	return nil
}

// GetFileInfo gets information about a file
func (s *StorageService) GetFileInfo(fileID, userID string) (string, string, int64, time.Time, error) {
	objectName := fmt.Sprintf("%s/%s", userID, fileID)

	ctx := context.Background()
	stat, err := s.minioClient.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", "", 0, time.Time{}, fmt.Errorf("failed to stat object: %w", err)
	}

	return fileID, stat.ContentType, stat.Size, stat.LastModified, nil
}
