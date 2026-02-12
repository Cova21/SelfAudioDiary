package grpc

import (
	"context"

	"github.com/voice-diary/backend/search-service/service"
	searchpb "github.com/voice-diary/backend/internal/gen/search"
)

type searchHandler struct {
	searchpb.UnimplementedSearchServiceServer
	searchService *service.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *service.SearchService) searchpb.SearchServiceServer {
	return &searchHandler{
		searchService: searchService,
	}
}

func (h *searchHandler) IndexEntry(ctx context.Context, req *searchpb.IndexEntryRequest) (*searchpb.IndexEntryResponse, error) {
	doc := &service.DiaryEntryDocument{
		EntryID:       req.EntryId,
		UserID:        req.UserId,
		Title:         req.Title,
		Transcription: req.Transcription,
		Sentiment:     req.Sentiment,
		Emotions:      req.Emotions,
		Keywords:      req.Keywords,
		Tags:          req.Tags,
		CreatedAt:     req.CreatedAt,
	}

	err := h.searchService.IndexEntry(doc)
	if err != nil {
		return &searchpb.IndexEntryResponse{Success: false}, err
	}

	return &searchpb.IndexEntryResponse{Success: true}, nil
}

func (h *searchHandler) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchResponse, error) {
	results, total, err := h.searchService.Search(
		req.UserId,
		req.Query,
		req.Emotions,
		req.Tags,
		req.Sentiment,
		req.DateFrom,
		req.DateTo,
		int(req.Page),
		int(req.PageSize),
	)
	if err != nil {
		return nil, err
	}

	pbResults := make([]*searchpb.SearchResult, len(results))
	for i, result := range results {
		pbResults[i] = &searchpb.SearchResult{
			EntryId:              result.EntryID,
			Title:                result.Title,
			TranscriptionSnippet: result.TranscriptionSnippet,
			Highlights:           result.Highlights,
			Score:                float32(result.Score),
			CreatedAt:            result.CreatedAt,
		}
	}

	return &searchpb.SearchResponse{
		Results:  pbResults,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *searchHandler) DeleteEntry(ctx context.Context, req *searchpb.DeleteEntryRequest) (*searchpb.DeleteEntryResponse, error) {
	err := h.searchService.DeleteEntry(req.EntryId)
	if err != nil {
		return &searchpb.DeleteEntryResponse{Success: false}, err
	}

	return &searchpb.DeleteEntryResponse{Success: true}, nil
}

func (h *searchHandler) UpdateEntry(ctx context.Context, req *searchpb.UpdateEntryRequest) (*searchpb.UpdateEntryResponse, error) {
	doc := &service.DiaryEntryDocument{
		EntryID:       req.EntryId,
		UserID:        req.UserId,
		Title:         req.Title,
		Transcription: req.Transcription,
		Tags:          req.Tags,
	}

	err := h.searchService.UpdateEntry(doc)
	if err != nil {
		return &searchpb.UpdateEntryResponse{Success: false}, err
	}

	return &searchpb.UpdateEntryResponse{Success: true}, nil
}
