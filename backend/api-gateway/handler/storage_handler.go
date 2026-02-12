package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	storagepb "github.com/voice-diary/backend/internal/gen/storage"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type StorageHandler struct {
	storageClient storagepb.StorageServiceClient
}

func NewStorageHandler(storageConn *grpc.ClientConn) *StorageHandler {
	return &StorageHandler{
		storageClient: storagepb.NewStorageServiceClient(storageConn),
	}
}

type GenerateUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

func (h *StorageHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req GenerateUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.storageClient.GenerateUploadURL(ctx, &storagepb.GenerateUploadURLRequest{
		UserId:      userID,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
	})

	if err != nil {
		logger.Error("Generate upload URL failed", zap.Error(err))
		http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp)
}

func (h *StorageHandler) GenerateDownloadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["id"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.storageClient.GenerateDownloadURL(ctx, &storagepb.GenerateDownloadURLRequest{
		FileId: fileID,
		UserId: userID,
	})

	if err != nil {
		logger.Error("Generate download URL failed", zap.Error(err))
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp)
}

func (h *StorageHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["id"]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := h.storageClient.DownloadFile(ctx, &storagepb.DownloadFileRequest{
		FileId: fileID,
		UserId: userID,
	})

	if err != nil {
		logger.Error("Download file failed", zap.Error(err))
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+resp.Filename+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(resp.FileSize, 10))
	w.Header().Set("Accept-Ranges", "bytes")
	
	// Write file content
	w.Write(resp.Content)
}

func (h *StorageHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["id"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := h.storageClient.DeleteFile(ctx, &storagepb.DeleteFileRequest{
		FileId: fileID,
		UserId: userID,
	})

	if err != nil {
		logger.Error("Delete file failed", zap.Error(err))
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
