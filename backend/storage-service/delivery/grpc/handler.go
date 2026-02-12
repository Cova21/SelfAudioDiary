package grpc

import (
	"context"

	"github.com/voice-diary/backend/storage-service/service"
	storagepb "github.com/voice-diary/backend/internal/gen/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type storageHandler struct {
	storagepb.UnimplementedStorageServiceServer
	storageService *service.StorageService
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(storageService *service.StorageService) storagepb.StorageServiceServer {
	return &storageHandler{
		storageService: storageService,
	}
}

func (h *storageHandler) GenerateUploadURL(ctx context.Context, req *storagepb.GenerateUploadURLRequest) (*storagepb.GenerateUploadURLResponse, error) {
	uploadURL, fileID, err := h.storageService.GenerateUploadURL(
		req.UserId,
		req.Filename,
		req.ContentType,
		req.FileSize,
	)
	if err != nil {
		return nil, err
	}

	return &storagepb.GenerateUploadURLResponse{
		UploadUrl: uploadURL,
		FileId:    fileID,
		ExpiresIn: 900, // 15 minutes
	}, nil
}

func (h *storageHandler) GenerateDownloadURL(ctx context.Context, req *storagepb.GenerateDownloadURLRequest) (*storagepb.GenerateDownloadURLResponse, error) {
	downloadURL, err := h.storageService.GenerateDownloadURL(req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &storagepb.GenerateDownloadURLResponse{
		DownloadUrl: downloadURL,
		ExpiresIn:   3600, // 1 hour
	}, nil
}

func (h *storageHandler) UploadFile(ctx context.Context, req *storagepb.UploadFileRequest) (*storagepb.UploadFileResponse, error) {
	fileID, url, fileSize, err := h.storageService.UploadFile(
		req.UserId,
		req.Filename,
		req.Content,
		req.ContentType,
	)
	if err != nil {
		return nil, err
	}

	return &storagepb.UploadFileResponse{
		FileId:   fileID,
		Url:      url,
		FileSize: fileSize,
	}, nil
}

func (h *storageHandler) DownloadFile(ctx context.Context, req *storagepb.DownloadFileRequest) (*storagepb.DownloadFileResponse, error) {
	content, filename, contentType, fileSize, err := h.storageService.DownloadFile(req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &storagepb.DownloadFileResponse{
		Content:     content,
		Filename:    filename,
		ContentType: contentType,
		FileSize:    fileSize,
	}, nil
}

func (h *storageHandler) DeleteFile(ctx context.Context, req *storagepb.DeleteFileRequest) (*storagepb.DeleteFileResponse, error) {
	err := h.storageService.DeleteFile(req.FileId, req.UserId)
	if err != nil {
		return &storagepb.DeleteFileResponse{Success: false}, err
	}

	return &storagepb.DeleteFileResponse{Success: true}, nil
}

func (h *storageHandler) GetFileInfo(ctx context.Context, req *storagepb.GetFileInfoRequest) (*storagepb.GetFileInfoResponse, error) {
	fileID, contentType, fileSize, createdAt, err := h.storageService.GetFileInfo(req.FileId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &storagepb.GetFileInfoResponse{
		FileInfo: &storagepb.FileInfo{
			FileId:      fileID,
			UserId:      req.UserId,
			Filename:    fileID,
			ContentType: contentType,
			FileSize:    fileSize,
			CreatedAt:   timestamppb.New(createdAt),
		},
	}, nil
}
