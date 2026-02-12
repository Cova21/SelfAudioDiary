package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

const indexName = "diary_entries"

// DiaryEntryDocument represents a diary entry in Elasticsearch
type DiaryEntryDocument struct {
	EntryID       string   `json:"entry_id"`
	UserID        string   `json:"user_id"`
	Title         string   `json:"title"`
	Transcription string   `json:"transcription"`
	Sentiment     string   `json:"sentiment"`
	Emotions      []string `json:"emotions"`
	Keywords      []string `json:"keywords"`
	Tags          []string `json:"tags"`
	CreatedAt     int64    `json:"created_at"`
}

// SearchResult represents a search result
type SearchResult struct {
	EntryID              string
	Title                string
	TranscriptionSnippet string
	Highlights           []string
	Score                float64
	CreatedAt            int64
}

// SearchService handles search operations
type SearchService struct {
	esClient *elastic.Client
}

// NewSearchService creates a new search service
func NewSearchService(elasticsearchURL string) (*SearchService, error) {
	// Create Elasticsearch client
	client, err := elastic.NewClient(
		elastic.SetURL(elasticsearchURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	logger.Info("Connected to Elasticsearch", zap.String("url", elasticsearchURL))

	// Create index if it doesn't exist
	ctx := context.Background()
	exists, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		mapping := `{
			"mappings": {
				"properties": {
					"entry_id": {"type": "keyword"},
					"user_id": {"type": "keyword"},
					"title": {"type": "text"},
					"transcription": {"type": "text"},
					"sentiment": {"type": "keyword"},
					"emotions": {"type": "keyword"},
					"keywords": {"type": "keyword"},
					"tags": {"type": "keyword"},
					"created_at": {"type": "long"}
				}
			}
		}`

		_, err = client.CreateIndex(indexName).BodyString(mapping).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
		logger.Info("Created Elasticsearch index", zap.String("index", indexName))
	}

	return &SearchService{
		esClient: client,
	}, nil
}

// IndexEntry indexes a diary entry
func (s *SearchService) IndexEntry(doc *DiaryEntryDocument) error {
	ctx := context.Background()

	_, err := s.esClient.Index().
		Index(indexName).
		Id(doc.EntryID).
		BodyJson(doc).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index entry: %w", err)
	}

	logger.Info("Indexed entry", zap.String("entry_id", doc.EntryID))
	return nil
}

// Search searches for diary entries
func (s *SearchService) Search(
	userID, query string,
	emotions, tags []string,
	sentiment string,
	dateFrom, dateTo int64,
	page, pageSize int,
) ([]*SearchResult, int, error) {
	ctx := context.Background()

	// Build query
	boolQuery := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("user_id", userID))

	// Add text search
	if query != "" {
		multiMatchQuery := elastic.NewMultiMatchQuery(query, "title", "transcription").
			Type("best_fields").
			Fuzziness("AUTO")
		boolQuery = boolQuery.Must(multiMatchQuery)
	}

	// Add filters
	if len(emotions) > 0 {
		boolQuery = boolQuery.Filter(elastic.NewTermsQuery("emotions", stringsToInterfaces(emotions)...))
	}

	if len(tags) > 0 {
		boolQuery = boolQuery.Filter(elastic.NewTermsQuery("tags", stringsToInterfaces(tags)...))
	}

	if sentiment != "" {
		boolQuery = boolQuery.Filter(elastic.NewTermQuery("sentiment", sentiment))
	}

	if dateFrom > 0 || dateTo > 0 {
		rangeQuery := elastic.NewRangeQuery("created_at")
		if dateFrom > 0 {
			rangeQuery = rangeQuery.Gte(dateFrom)
		}
		if dateTo > 0 {
			rangeQuery = rangeQuery.Lte(dateTo)
		}
		boolQuery = boolQuery.Filter(rangeQuery)
	}

	// Calculate offset
	from := (page - 1) * pageSize

	// Execute search
	searchResult, err := s.esClient.Search().
		Index(indexName).
		Query(boolQuery).
		Highlight(
			elastic.NewHighlight().
				Field("transcription").
				PreTags("<mark>").
				PostTags("</mark>"),
		).
		From(from).
		Size(pageSize).
		Sort("created_at", false).
		Do(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to search: %w", err)
	}

	// Parse results
	results := make([]*SearchResult, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var doc DiaryEntryDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logger.Error("Failed to unmarshal search result", zap.Error(err))
			continue
		}

		// Extract highlights
		highlights := []string{}
		if hit.Highlight != nil {
			if transcriptionHighlights, ok := hit.Highlight["transcription"]; ok {
				highlights = transcriptionHighlights
			}
		}

		// Create snippet from transcription
		snippet := doc.Transcription
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		results = append(results, &SearchResult{
			EntryID:              doc.EntryID,
			Title:                doc.Title,
			TranscriptionSnippet: snippet,
			Highlights:           highlights,
			Score:                float64(*hit.Score),
			CreatedAt:            doc.CreatedAt,
		})
	}

	total := int(searchResult.Hits.TotalHits.Value)

	logger.Info("Search completed",
		zap.String("user_id", userID),
		zap.String("query", query),
		zap.Int("results", len(results)),
		zap.Int("total", total),
	)

	return results, total, nil
}

// DeleteEntry deletes an entry from the index
func (s *SearchService) DeleteEntry(entryID string) error {
	ctx := context.Background()

	_, err := s.esClient.Delete().
		Index(indexName).
		Id(entryID).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	logger.Info("Deleted entry from index", zap.String("entry_id", entryID))
	return nil
}

// UpdateEntry updates an entry in the index
func (s *SearchService) UpdateEntry(doc *DiaryEntryDocument) error {
	// Simply re-index the document
	return s.IndexEntry(doc)
}

// Helper function to convert strings to interfaces
func stringsToInterfaces(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}
