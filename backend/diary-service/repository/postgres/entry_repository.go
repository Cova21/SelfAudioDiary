package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/voice-diary/backend/diary-service/domain"
)

type entryRepository struct {
	db *sql.DB
}

// NewEntryRepository creates a new entry repository
func NewEntryRepository(db *sql.DB) domain.EntryRepository {
	return &entryRepository{db: db}
}

func (r *entryRepository) Create(entry *domain.DiaryEntry) error {
	entry.ID = uuid.New().String()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()
	entry.Status = domain.StatusUploaded

	segmentsJSON, err := json.Marshal(entry.Segments)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %w", err)
	}

	query := `
		INSERT INTO diary_entries (
			id, user_id, title, audio_file_id, duration_seconds,
			transcription, segments, language, sentiment, emotions,
			keywords, tags, summary, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = r.db.Exec(
		query,
		entry.ID, entry.UserID, entry.Title, entry.AudioFileID, entry.DurationSeconds,
		entry.Transcription, segmentsJSON, entry.Language, entry.Sentiment,
		pq.Array(entry.Emotions), pq.Array(entry.Keywords), pq.Array(entry.Tags),
		entry.Summary, entry.Status, entry.CreatedAt, entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

func (r *entryRepository) GetByID(id, userID string) (*domain.DiaryEntry, error) {
	query := `
		SELECT id, user_id, title, audio_file_id, duration_seconds,
			transcription, segments, language, sentiment, emotions,
			keywords, tags, summary, status, created_at, updated_at
		FROM diary_entries
		WHERE id = $1 AND user_id = $2
	`

	entry := &domain.DiaryEntry{}
	var segmentsJSON []byte

	err := r.db.QueryRow(query, id, userID).Scan(
		&entry.ID, &entry.UserID, &entry.Title, &entry.AudioFileID, &entry.DurationSeconds,
		&entry.Transcription, &segmentsJSON, &entry.Language, &entry.Sentiment,
		pq.Array(&entry.Emotions), pq.Array(&entry.Keywords), pq.Array(&entry.Tags),
		&entry.Summary, &entry.Status, &entry.CreatedAt, &entry.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("entry not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Unmarshal segments
	if len(segmentsJSON) > 0 {
		if err := json.Unmarshal(segmentsJSON, &entry.Segments); err != nil {
			return nil, fmt.Errorf("failed to unmarshal segments: %w", err)
		}
	}

	return entry, nil
}

func (r *entryRepository) List(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
	// Validate sort parameters
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM diary_entries WHERE user_id = $1`
	if err := r.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count entries: %w", err)
	}

	// Get entries
	query := fmt.Sprintf(`
		SELECT id, user_id, title, audio_file_id, duration_seconds,
			transcription, segments, language, sentiment, emotions,
			keywords, tags, summary, status, created_at, updated_at
		FROM diary_entries
		WHERE user_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, sortOrder)

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	entries := []*domain.DiaryEntry{}
	for rows.Next() {
		entry := &domain.DiaryEntry{}
		var segmentsJSON []byte

		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Title, &entry.AudioFileID, &entry.DurationSeconds,
			&entry.Transcription, &segmentsJSON, &entry.Language, &entry.Sentiment,
			pq.Array(&entry.Emotions), pq.Array(&entry.Keywords), pq.Array(&entry.Tags),
			&entry.Summary, &entry.Status, &entry.CreatedAt, &entry.UpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan entry: %w", err)
		}

		// Unmarshal segments
		if len(segmentsJSON) > 0 {
			if err := json.Unmarshal(segmentsJSON, &entry.Segments); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal segments: %w", err)
			}
		}

		entries = append(entries, entry)
	}

	return entries, total, nil
}

func (r *entryRepository) Update(entry *domain.DiaryEntry) error {
	entry.UpdatedAt = time.Now()

	segmentsJSON, err := json.Marshal(entry.Segments)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %w", err)
	}

	query := `
		UPDATE diary_entries
		SET title = $1, transcription = $2, segments = $3, language = $4,
		    sentiment = $5, emotions = $6, keywords = $7, tags = $8, 
		    summary = $9, status = $10, updated_at = $11
		WHERE id = $12 AND user_id = $13
	`

	_, err = r.db.Exec(query, 
		entry.Title, entry.Transcription, segmentsJSON, entry.Language,
		entry.Sentiment, pq.Array(entry.Emotions), pq.Array(entry.Keywords), pq.Array(entry.Tags),
		entry.Summary, entry.Status, entry.UpdatedAt, entry.ID, entry.UserID)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

func (r *entryRepository) Delete(id, userID string) error {
	query := `DELETE FROM diary_entries WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("entry not found")
	}

	return nil
}

func (r *entryRepository) UpdateTranscription(id string, transcription string, segments []domain.TranscriptionSegment, language string) error {
	segmentsJSON, err := json.Marshal(segments)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %w", err)
	}

	query := `
		UPDATE diary_entries
		SET transcription = $1, segments = $2, language = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	_, err = r.db.Exec(
		query,
		transcription, segmentsJSON, language, domain.StatusTranscribed, time.Now(), id,
	)

	if err != nil {
		return fmt.Errorf("failed to update transcription: %w", err)
	}

	return nil
}

func (r *entryRepository) UpdateAnalysis(id, sentiment, summary string, emotions, keywords []string) error {
	query := `
		UPDATE diary_entries
		SET sentiment = $1, summary = $2, emotions = $3, keywords = $4, status = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(
		query,
		sentiment, summary, pq.Array(emotions), pq.Array(keywords), domain.StatusCompleted, time.Now(), id,
	)

	if err != nil {
		return fmt.Errorf("failed to update analysis: %w", err)
	}

	return nil
}
