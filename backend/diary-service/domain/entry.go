package domain

import (
	"time"
)

// ProcessingStatus represents the status of entry processing
type ProcessingStatus int

const (
	StatusUploaded ProcessingStatus = iota
	StatusTranscribing
	StatusTranscribed
	StatusAnalyzing
	StatusCompleted
	StatusFailed
)

// DiaryEntry represents a diary entry entity
type DiaryEntry struct {
	ID              string
	UserID          string
	Title           string
	AudioFileID     string
	DurationSeconds int32
	Transcription   string
	Segments        []TranscriptionSegment
	Language        string
	Sentiment       string
	Emotions        []string
	Keywords        []string
	Tags            []string
	Summary         string
	Status          ProcessingStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TranscriptionSegment represents a segment of transcription with timing
type TranscriptionSegment struct {
	StartTimeMs int32
	EndTimeMs   int32
	Text        string
}

// EntryRepository defines the interface for diary entry operations
type EntryRepository interface {
	Create(entry *DiaryEntry) error
	GetByID(id, userID string) (*DiaryEntry, error)
	List(userID string, page, pageSize int, sortBy, sortOrder string) ([]*DiaryEntry, int, error)
	Update(entry *DiaryEntry) error
	Delete(id, userID string) error
	UpdateTranscription(id string, transcription string, segments []TranscriptionSegment, language string) error
	UpdateAnalysis(id, sentiment, summary string, emotions, keywords []string) error
}
