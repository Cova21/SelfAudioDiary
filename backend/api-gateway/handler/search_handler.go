package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	searchpb "github.com/voice-diary/backend/internal/gen/search"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type SearchHandler struct {
	searchClient searchpb.SearchServiceClient
}

func NewSearchHandler(searchConn *grpc.ClientConn) *SearchHandler {
	return &SearchHandler{
		searchClient: searchpb.NewSearchServiceClient(searchConn),
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	emotions := r.URL.Query()["emotions"]
	tags := r.URL.Query()["tags"]
	sentiment := r.URL.Query().Get("sentiment")

	dateFrom, _ := strconv.ParseInt(r.URL.Query().Get("date_from"), 10, 64)
	dateTo, _ := strconv.ParseInt(r.URL.Query().Get("date_to"), 10, 64)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.searchClient.Search(ctx, &searchpb.SearchRequest{
		UserId:    userID,
		Query:     query,
		Emotions:  emotions,
		Tags:      tags,
		Sentiment: sentiment,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Page:      int32(page),
		PageSize:  int32(pageSize),
	})

	if err != nil {
		logger.Error("Search failed", zap.Error(err))
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
