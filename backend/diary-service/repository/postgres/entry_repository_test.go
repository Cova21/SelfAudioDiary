package postgres

import (
	"testing"

	"github.com/voice-diary/backend/diary-service/domain"
)

func TestNewEntryRepository_NotNil(t *testing.T) {
	// Arrange/Act: создаём repository записей дневника.
	repo := NewEntryRepository(nil)
	// Assert: репозиторий должен быть создан.
	if repo == nil {
		t.Fatal("expected non-nil entry repository")
	}
	// Assert: проверяем соответствие интерфейсу EntryRepository.
	var _ domain.EntryRepository = repo
}
