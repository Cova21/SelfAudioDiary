package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/voice-diary/backend/diary-service/domain"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/rabbitmq"
	"go.uber.org/zap"
)

// DiaryService handles diary entry business logic
type DiaryService struct {
	entryRepo domain.EntryRepository
	mq        *rabbitmq.Client
}

// NewDiaryService creates a new diary service
func NewDiaryService(entryRepo domain.EntryRepository, mq *rabbitmq.Client) *DiaryService {
	return &DiaryService{
		entryRepo: entryRepo,
		mq:        mq,
	}
}

// CreateEntry creates a new diary entry
func (s *DiaryService) CreateEntry(userID, title, audioFileID string, durationSeconds int32) (*domain.DiaryEntry, error) {
	entry := &domain.DiaryEntry{
		UserID:          userID,
		Title:           title,
		AudioFileID:     audioFileID,
		DurationSeconds: durationSeconds,
		Status:          domain.StatusUploaded,
	}

	if err := s.entryRepo.Create(entry); err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	// Send to transcription queue
	if err := s.sendToTranscriptionQueue(entry); err != nil {
		logger.Error("Failed to send to transcription queue", zap.Error(err))
		// Don't fail the request, just log the error
	}

	logger.Info("Created diary entry", zap.String("entry_id", entry.ID), zap.String("user_id", userID))

	return entry, nil
}

// GetEntry retrieves a diary entry by ID
func (s *DiaryService) GetEntry(id, userID string) (*domain.DiaryEntry, error) {
	entry, err := s.entryRepo.GetByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	return entry, nil
}

// ListEntries retrieves a list of diary entries
func (s *DiaryService) ListEntries(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entries, total, err := s.entryRepo.List(userID, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list entries: %w", err)
	}

	return entries, total, nil
}

// UpdateEntry updates a diary entry
func (s *DiaryService) UpdateEntry(id, userID, title string, tags []string) (*domain.DiaryEntry, error) {
	// Get existing entry
	entry, err := s.entryRepo.GetByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Update fields
	entry.Title = title
	entry.Tags = tags

	if err := s.entryRepo.Update(entry); err != nil {
		return nil, fmt.Errorf("failed to update entry: %w", err)
	}

	logger.Info("Updated diary entry", zap.String("entry_id", id), zap.String("user_id", userID))

	return entry, nil
}

// DeleteEntry deletes a diary entry
func (s *DiaryService) DeleteEntry(id, userID string) error {
	if err := s.entryRepo.Delete(id, userID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	logger.Info("Deleted diary entry", zap.String("entry_id", id), zap.String("user_id", userID))

	return nil
}

// UpdateTranscription updates the transcription of an entry (called by transcription-service)
func (s *DiaryService) UpdateTranscription(entryID, transcription, language string, segments []domain.TranscriptionSegment) error {
	if err := s.entryRepo.UpdateTranscription(entryID, transcription, segments, language); err != nil {
		return fmt.Errorf("failed to update transcription: %w", err)
	}

	logger.Info("Updated transcription", zap.String("entry_id", entryID))

	// Send to NLP analysis queue
	if err := s.sendToNLPQueue(entryID, transcription); err != nil {
		logger.Error("Failed to send to NLP queue", zap.Error(err))
	}

	return nil
}

// UpdateAnalysis updates the analysis of an entry (called by nlp-service)
func (s *DiaryService) UpdateAnalysis(entryID, sentiment, summary string, emotions, keywords []string) error {
	if err := s.entryRepo.UpdateAnalysis(entryID, sentiment, summary, emotions, keywords); err != nil {
		return fmt.Errorf("failed to update analysis: %w", err)
	}

	logger.Info("Updated analysis", zap.String("entry_id", entryID))

	return nil
}

// sendToTranscriptionQueue sends entry to transcription queue
func (s *DiaryService) sendToTranscriptionQueue(entry *domain.DiaryEntry) error {
	message := map[string]interface{}{
		"entry_id":      entry.ID,
		"user_id":       entry.UserID,
		"audio_file_id": entry.AudioFileID,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Declare queue if not exists
	if err := s.mq.DeclareQueue("transcription_queue"); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	ctx := context.Background()
	if err := s.mq.Publish(ctx, "transcription_queue", messageJSON); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Info("Sent to transcription queue", zap.String("entry_id", entry.ID))

	return nil
}

// sendToNLPQueue sends entry to NLP analysis queue
func (s *DiaryService) sendToNLPQueue(entryID, transcription string) error {
	message := map[string]interface{}{
		"entry_id":      entryID,
		"transcription": transcription,
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Declare queue if not exists
	if err := s.mq.DeclareQueue("nlp_queue"); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	ctx := context.Background()
	if err := s.mq.Publish(ctx, "nlp_queue", messageJSON); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Info("Sent to NLP queue", zap.String("entry_id", entryID))

	return nil
}

// StartNLPConsumer starts consuming NLP analysis results
func (s *DiaryService) StartNLPConsumer() error {
	queueName := "nlp_completed"

	if err := s.mq.DeclareQueue(queueName); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Info("Starting NLP consumer", zap.String("queue", queueName))

	return s.mq.Consume(queueName, func(body []byte) error {
		var message struct {
			EntryID   string   `json:"entry_id"`
			Sentiment string   `json:"sentiment"`
			Emotions  []string `json:"emotions"`
			Keywords  []string `json:"keywords"`
			Summary   string   `json:"summary"`
		}

		if err := json.Unmarshal(body, &message); err != nil {
			logger.Error("Failed to unmarshal NLP message", zap.Error(err))
			return err
		}

		logger.Info("Received NLP result",
			zap.String("entry_id", message.EntryID),
			zap.String("sentiment", message.Sentiment),
			zap.Strings("emotions", message.Emotions))

		if err := s.entryRepo.UpdateAnalysis(message.EntryID, message.Sentiment, message.Summary, message.Emotions, message.Keywords); err != nil {
			logger.Error("Failed to update entry with analysis", zap.Error(err))
			return err
		}

		logger.Info("Updated entry with NLP analysis", zap.String("entry_id", message.EntryID))

		return nil
	})
}

// StartTranscriptionConsumer starts consuming transcription results
func (s *DiaryService) StartTranscriptionConsumer() error {
	queueName := "transcription_completed"
	
	// Declare queue if not exists
	if err := s.mq.DeclareQueue(queueName); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Info("Starting transcription consumer", zap.String("queue", queueName))

	// Consume messages
	return s.mq.Consume(queueName, func(body []byte) error {
		var message struct {
			EntryID       string `json:"entry_id"`
			UserID        string `json:"user_id"`
			Transcription string `json:"transcription"`
			Language      string `json:"language"`
		}

		if err := json.Unmarshal(body, &message); err != nil {
			logger.Error("Failed to unmarshal transcription message", zap.Error(err))
			return err
		}

		logger.Info("Received transcription result", 
			zap.String("entry_id", message.EntryID),
			zap.Int("transcription_length", len(message.Transcription)))

		// Update entry with transcription
		entry, err := s.entryRepo.GetByID(message.EntryID, message.UserID)
		if err != nil {
			logger.Error("Failed to get entry for transcription update", zap.Error(err))
			return err
		}

		entry.Transcription = message.Transcription
		entry.Language = message.Language
		entry.Status = domain.StatusTranscribed

		if err := s.entryRepo.Update(entry); err != nil {
			logger.Error("Failed to update entry with transcription", zap.Error(err))
			return err
		}

		logger.Info("Updated entry with transcription", zap.String("entry_id", message.EntryID))

		// Send to NLP queue for analysis
		if err := s.sendToNLPQueue(message.EntryID, message.Transcription); err != nil {
			logger.Error("Failed to send to NLP queue", zap.Error(err))
			// Don't fail the message, just log
		}

		return nil
	})
}
