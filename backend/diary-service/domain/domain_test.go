package domain

import "testing"

func TestProcessingStatusOrder(t *testing.T) {
	// Assert: статусы обработки должны идти в ожидаемом порядке.
	if !(StatusUploaded < StatusTranscribing &&
		StatusTranscribing < StatusTranscribed &&
		StatusTranscribed < StatusAnalyzing &&
		StatusAnalyzing < StatusCompleted) {
		t.Fatal("unexpected processing status order")
	}
}

func TestDiaryEntryFields(t *testing.T) {
	// Arrange: создаём тестовую запись дневника.
	e := DiaryEntry{ID: "e1", UserID: "u1", Title: "title"}
	// Assert: обязательные поля записи должны быть заполнены.
	if e.ID == "" || e.UserID == "" || e.Title == "" {
		t.Fatal("required diary entry fields should not be empty")
	}
}
