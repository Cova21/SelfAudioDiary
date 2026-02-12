package grpc

import (
	"context"

	"github.com/voice-diary/backend/diary-service/domain"
	"github.com/voice-diary/backend/diary-service/service"
	diarypb "github.com/voice-diary/backend/internal/gen/diary"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type diaryHandler struct {
	diarypb.UnimplementedDiaryServiceServer
	diaryService *service.DiaryService
}

// NewDiaryHandler creates a new diary handler
func NewDiaryHandler(diaryService *service.DiaryService) diarypb.DiaryServiceServer {
	return &diaryHandler{
		diaryService: diaryService,
	}
}

func (h *diaryHandler) CreateEntry(ctx context.Context, req *diarypb.CreateEntryRequest) (*diarypb.CreateEntryResponse, error) {
	entry, err := h.diaryService.CreateEntry(req.UserId, req.Title, req.AudioFileId, req.DurationSeconds)
	if err != nil {
		return nil, err
	}

	return &diarypb.CreateEntryResponse{
		Entry: convertToPB(entry),
	}, nil
}

func (h *diaryHandler) GetEntry(ctx context.Context, req *diarypb.GetEntryRequest) (*diarypb.GetEntryResponse, error) {
	entry, err := h.diaryService.GetEntry(req.EntryId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &diarypb.GetEntryResponse{
		Entry: convertToPB(entry),
	}, nil
}

func (h *diaryHandler) ListEntries(ctx context.Context, req *diarypb.ListEntriesRequest) (*diarypb.ListEntriesResponse, error) {
	entries, total, err := h.diaryService.ListEntries(
		req.UserId,
		int(req.Page),
		int(req.PageSize),
		req.SortBy,
		req.SortOrder,
	)
	if err != nil {
		return nil, err
	}

	pbEntries := make([]*diarypb.DiaryEntry, len(entries))
	for i, entry := range entries {
		pbEntries[i] = convertToPB(entry)
	}

	return &diarypb.ListEntriesResponse{
		Entries:  pbEntries,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *diaryHandler) UpdateEntry(ctx context.Context, req *diarypb.UpdateEntryRequest) (*diarypb.UpdateEntryResponse, error) {
	entry, err := h.diaryService.UpdateEntry(req.EntryId, req.UserId, req.Title, req.Tags)
	if err != nil {
		return nil, err
	}

	return &diarypb.UpdateEntryResponse{
		Entry: convertToPB(entry),
	}, nil
}

func (h *diaryHandler) DeleteEntry(ctx context.Context, req *diarypb.DeleteEntryRequest) (*diarypb.DeleteEntryResponse, error) {
	err := h.diaryService.DeleteEntry(req.EntryId, req.UserId)
	if err != nil {
		return &diarypb.DeleteEntryResponse{Success: false}, err
	}

	return &diarypb.DeleteEntryResponse{Success: true}, nil
}

func (h *diaryHandler) UpdateTranscription(ctx context.Context, req *diarypb.UpdateTranscriptionRequest) (*diarypb.UpdateTranscriptionResponse, error) {
	segments := make([]domain.TranscriptionSegment, len(req.Segments))
	for i, seg := range req.Segments {
		segments[i] = domain.TranscriptionSegment{
			StartTimeMs: seg.StartTimeMs,
			EndTimeMs:   seg.EndTimeMs,
			Text:        seg.Text,
		}
	}

	err := h.diaryService.UpdateTranscription(req.EntryId, req.Transcription, req.Language, segments)
	if err != nil {
		return &diarypb.UpdateTranscriptionResponse{Success: false}, err
	}

	return &diarypb.UpdateTranscriptionResponse{Success: true}, nil
}

func (h *diaryHandler) UpdateAnalysis(ctx context.Context, req *diarypb.UpdateAnalysisRequest) (*diarypb.UpdateAnalysisResponse, error) {
	err := h.diaryService.UpdateAnalysis(req.EntryId, req.Sentiment, req.Summary, req.Emotions, req.Keywords)
	if err != nil {
		return &diarypb.UpdateAnalysisResponse{Success: false}, err
	}

	return &diarypb.UpdateAnalysisResponse{Success: true}, nil
}

// Helper function to convert domain entry to protobuf
func convertToPB(entry *domain.DiaryEntry) *diarypb.DiaryEntry {
	segments := make([]*diarypb.TranscriptionSegment, len(entry.Segments))
	for i, seg := range entry.Segments {
		segments[i] = &diarypb.TranscriptionSegment{
			StartTimeMs: seg.StartTimeMs,
			EndTimeMs:   seg.EndTimeMs,
			Text:        seg.Text,
		}
	}

	return &diarypb.DiaryEntry{
		Id:              entry.ID,
		UserId:          entry.UserID,
		Title:           entry.Title,
		AudioFileId:     entry.AudioFileID,
		DurationSeconds: entry.DurationSeconds,
		Transcription:   entry.Transcription,
		Segments:        segments,
		Language:        entry.Language,
		Sentiment:       entry.Sentiment,
		Emotions:        entry.Emotions,
		Keywords:        entry.Keywords,
		Tags:            entry.Tags,
		Summary:         entry.Summary,
		Status:          diarypb.ProcessingStatus(entry.Status),
		CreatedAt:       timestamppb.New(entry.CreatedAt),
		UpdatedAt:       timestamppb.New(entry.UpdatedAt),
	}
}
