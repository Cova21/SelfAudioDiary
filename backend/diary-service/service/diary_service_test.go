package service

import (
	"errors"
	"testing"

	"github.com/voice-diary/backend/diary-service/domain"
)

type mockEntryRepo struct {
	createFn              func(entry *domain.DiaryEntry) error
	getByIDFn             func(id, userID string) (*domain.DiaryEntry, error)
	listFn                func(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error)
	updateFn              func(entry *domain.DiaryEntry) error
	deleteFn              func(id, userID string) error
	updateTranscriptionFn func(id string, transcription string, segments []domain.TranscriptionSegment, language string) error
	updateAnalysisFn      func(id, sentiment, summary string, emotions, keywords []string) error
}

func (m *mockEntryRepo) Create(entry *domain.DiaryEntry) error {
	return m.createFn(entry)
}
func (m *mockEntryRepo) GetByID(id, userID string) (*domain.DiaryEntry, error) {
	return m.getByIDFn(id, userID)
}
func (m *mockEntryRepo) List(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
	return m.listFn(userID, page, pageSize, sortBy, sortOrder)
}
func (m *mockEntryRepo) Update(entry *domain.DiaryEntry) error {
	return m.updateFn(entry)
}
func (m *mockEntryRepo) Delete(id, userID string) error {
	return m.deleteFn(id, userID)
}
func (m *mockEntryRepo) UpdateTranscription(id string, transcription string, segments []domain.TranscriptionSegment, language string) error {
	return m.updateTranscriptionFn(id, transcription, segments, language)
}
func (m *mockEntryRepo) UpdateAnalysis(id, sentiment, summary string, emotions, keywords []string) error {
	return m.updateAnalysisFn(id, sentiment, summary, emotions, keywords)
}

func TestListEntries_NormalizesPageAndSize(t *testing.T) {
	// Arrange: подготавливаем репозиторий, который проверяет нормализованные page/pageSize.
	called := false
	repo := &mockEntryRepo{
		createFn:  func(entry *domain.DiaryEntry) error { return nil },
		getByIDFn: func(id, userID string) (*domain.DiaryEntry, error) { return nil, nil },
		listFn: func(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
			called = true
			if page != 1 {
				t.Fatalf("expected normalized page=1, got %d", page)
			}
			if pageSize != 20 {
				t.Fatalf("expected normalized pageSize=20, got %d", pageSize)
			}
			return []*domain.DiaryEntry{{ID: "e1"}}, 1, nil
		},
		updateFn:              func(entry *domain.DiaryEntry) error { return nil },
		deleteFn:              func(id, userID string) error { return nil },
		updateTranscriptionFn: func(id string, transcription string, segments []domain.TranscriptionSegment, language string) error { return nil },
		updateAnalysisFn:      func(id, sentiment, summary string, emotions, keywords []string) error { return nil },
	}
	svc := NewDiaryService(repo, nil)

	// Act: вызываем ListEntries с невалидными параметрами пагинации.
	entries, total, err := svc.ListEntries("u1", 0, 999, "created_at", "desc")
	// Assert: проверяем, что вызов успешен и репозиторий был вызван.
	if err != nil {
		t.Fatalf("ListEntries() error: %v", err)
	}
	if !called {
		t.Fatal("expected repository List() to be called")
	}
	if len(entries) != 1 || total != 1 {
		t.Fatalf("unexpected result: len=%d total=%d", len(entries), total)
	}
}

func TestUpdateEntry(t *testing.T) {
	// Arrange: подготавливаем существующую запись и мок Update.
	repo := &mockEntryRepo{
		createFn: func(entry *domain.DiaryEntry) error { return nil },
		getByIDFn: func(id, userID string) (*domain.DiaryEntry, error) {
			return &domain.DiaryEntry{ID: id, UserID: userID, Title: "old", Tags: []string{"old"}}, nil
		},
		listFn: func(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
			return nil, 0, nil
		},
		updateFn: func(entry *domain.DiaryEntry) error {
			if entry.Title != "new title" {
				t.Fatalf("expected updated title, got %q", entry.Title)
			}
			if len(entry.Tags) != 2 {
				t.Fatalf("expected updated tags")
			}
			return nil
		},
		deleteFn:              func(id, userID string) error { return nil },
		updateTranscriptionFn: func(id string, transcription string, segments []domain.TranscriptionSegment, language string) error { return nil },
		updateAnalysisFn:      func(id, sentiment, summary string, emotions, keywords []string) error { return nil },
	}
	svc := NewDiaryService(repo, nil)

	// Act: обновляем запись.
	got, err := svc.UpdateEntry("e1", "u1", "new title", []string{"a", "b"})
	// Assert: обновление должно пройти, title должен измениться.
	if err != nil {
		t.Fatalf("UpdateEntry() error: %v", err)
	}
	if got.Title != "new title" {
		t.Fatalf("unexpected title: %q", got.Title)
	}
}

func TestDeleteEntry_PropagatesError(t *testing.T) {
	// Arrange: мок delete возвращает ошибку.
	repo := &mockEntryRepo{
		createFn:  func(entry *domain.DiaryEntry) error { return nil },
		getByIDFn: func(id, userID string) (*domain.DiaryEntry, error) { return nil, nil },
		listFn: func(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
			return nil, 0, nil
		},
		updateFn: func(entry *domain.DiaryEntry) error { return nil },
		deleteFn: func(id, userID string) error { return errors.New("db down") },
		updateTranscriptionFn: func(id string, transcription string, segments []domain.TranscriptionSegment, language string) error {
			return nil
		},
		updateAnalysisFn: func(id, sentiment, summary string, emotions, keywords []string) error { return nil },
	}
	svc := NewDiaryService(repo, nil)

	// Act: удаляем запись.
	err := svc.DeleteEntry("e1", "u1")
	// Assert: ошибка должна быть проброшена из репозитория.
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateAnalysis(t *testing.T) {
	// Arrange: отслеживаем вызов UpdateAnalysis в репозитории.
	called := false
	repo := &mockEntryRepo{
		createFn:  func(entry *domain.DiaryEntry) error { return nil },
		getByIDFn: func(id, userID string) (*domain.DiaryEntry, error) { return nil, nil },
		listFn: func(userID string, page, pageSize int, sortBy, sortOrder string) ([]*domain.DiaryEntry, int, error) {
			return nil, 0, nil
		},
		updateFn: func(entry *domain.DiaryEntry) error { return nil },
		deleteFn: func(id, userID string) error { return nil },
		updateTranscriptionFn: func(id string, transcription string, segments []domain.TranscriptionSegment, language string) error {
			return nil
		},
		updateAnalysisFn: func(id, sentiment, summary string, emotions, keywords []string) error {
			called = true
			if id != "e1" || sentiment != "positive" || summary == "" {
				t.Fatalf("unexpected update args")
			}
			return nil
		},
	}
	svc := NewDiaryService(repo, nil)

	// Act: передаём результат анализа.
	err := svc.UpdateAnalysis("e1", "positive", "sum", []string{"радость"}, []string{"проект"})
	// Assert: вызов должен пройти успешно и репозиторий должен быть вызван.
	if err != nil {
		t.Fatalf("UpdateAnalysis() error: %v", err)
	}
	if !called {
		t.Fatal("expected repository UpdateAnalysis() to be called")
	}
}
