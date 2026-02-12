package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	diarypb "github.com/voice-diary/backend/internal/gen/diary"
	storagepb "github.com/voice-diary/backend/internal/gen/storage"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type DiaryHandler struct {
	diaryClient   diarypb.DiaryServiceClient
	storageClient storagepb.StorageServiceClient
}

func NewDiaryHandler(diaryConn *grpc.ClientConn, storageConn *grpc.ClientConn) *DiaryHandler {
	return &DiaryHandler{
		diaryClient:   diarypb.NewDiaryServiceClient(diaryConn),
		storageClient: storagepb.NewStorageServiceClient(storageConn),
	}
}

type CreateEntryRequest struct {
	Title           string `json:"title"`
	AudioFileID     string `json:"audio_file_id"`
	DurationSeconds int32  `json:"duration_seconds"`
}

func (h *DiaryHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.diaryClient.CreateEntry(ctx, &diarypb.CreateEntryRequest{
		UserId:          userID,
		Title:           req.Title,
		AudioFileId:     req.AudioFileID,
		DurationSeconds: req.DurationSeconds,
	})

	if err != nil {
		logger.Error("Create entry failed", zap.Error(err))
		http.Error(w, "Failed to create entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp.Entry)
}

func (h *DiaryHandler) GetEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	entryID := vars["id"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.diaryClient.GetEntry(ctx, &diarypb.GetEntryRequest{
		EntryId: entryID,
		UserId:  userID,
	})

	if err != nil {
		logger.Error("Get entry failed", zap.Error(err))
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp.Entry)
}

func (h *DiaryHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := r.URL.Query().Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.diaryClient.ListEntries(ctx, &diarypb.ListEntriesRequest{
		UserId:    userID,
		Page:      int32(page),
		PageSize:  int32(pageSize),
		SortBy:    sortBy,
		SortOrder: sortOrder,
	})

	if err != nil {
		logger.Error("List entries failed", zap.Error(err))
		http.Error(w, "Failed to list entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp)
}

func (h *DiaryHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	entryID := vars["id"]

	var req struct {
		Title string   `json:"title"`
		Tags  []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.diaryClient.UpdateEntry(ctx, &diarypb.UpdateEntryRequest{
		EntryId: entryID,
		UserId:  userID,
		Title:   req.Title,
		Tags:    req.Tags,
	})

	if err != nil {
		logger.Error("Update entry failed", zap.Error(err))
		http.Error(w, "Failed to update entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeProtoJSON(w, resp.Entry)
}

func (h *DiaryHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	entryID := vars["id"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := h.diaryClient.DeleteEntry(ctx, &diarypb.DeleteEntryRequest{
		EntryId: entryID,
		UserId:  userID,
	})

	if err != nil {
		logger.Error("Delete entry failed", zap.Error(err))
		http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// UploadEntry handles audio file upload and creates a diary entry
func (h *DiaryHandler) UploadEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("audio")
	if err != nil {
		logger.Error("Failed to get file from form", zap.Error(err))
		http.Error(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get title and duration
	title := r.FormValue("title")
	if title == "" {
		title = fmt.Sprintf("Recording %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	durationStr := r.FormValue("duration_seconds")
	duration, _ := strconv.Atoi(durationStr)
	if duration <= 0 {
		duration = 0
	}

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read file", zap.Error(err))
		http.Error(w, "Failed to read audio file", http.StatusInternalServerError)
		return
	}

	// Upload to storage service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Default to audio/webm for recorded audio
		contentType = "audio/webm"
	} else if strings.Contains(contentType, ";") {
		// Remove codec information (e.g., "audio/webm;codecs=opus" -> "audio/webm")
		contentType = strings.Split(contentType, ";")[0]
	}

	uploadResp, err := h.storageClient.UploadFile(ctx, &storagepb.UploadFileRequest{
		UserId:      userID,
		Filename:    header.Filename,
		ContentType: contentType,
		Content:     fileData,
	})

	if err != nil {
		logger.Error("Failed to upload file to storage", zap.Error(err))
		http.Error(w, "Failed to upload audio file", http.StatusInternalServerError)
		return
	}

	// Create diary entry
	createResp, err := h.diaryClient.CreateEntry(ctx, &diarypb.CreateEntryRequest{
		UserId:          userID,
		Title:           title,
		AudioFileId:     uploadResp.FileId,
		DurationSeconds: int32(duration),
	})

	if err != nil {
		logger.Error("Failed to create diary entry", zap.Error(err))
		http.Error(w, "Failed to create diary entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeProtoJSON(w, createResp.Entry)
}

// Helper function to write protobuf messages as JSON
func writeProtoJSON(w http.ResponseWriter, msg proto.Message) {
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false,
	}
	
	data, err := marshaler.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal proto to JSON", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Write(data)
}
