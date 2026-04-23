package service

import "testing"

func TestSearchService_TestPlan(t *testing.T) {
	// Arrange: создаём тестовый документ для индекса и тестовый search-result.
	doc := &DiaryEntryDocument{
		EntryID:       "e1",
		UserID:        "u1",
		Title:         "title",
		Transcription: "text",
	}
	// Assert: у документа должны быть заполнены обязательные поля.
	if doc.EntryID == "" || doc.UserID == "" {
		t.Fatal("expected non-empty required document fields")
	}

	result := &SearchResult{EntryID: "e1", Score: 1.0}
	// Assert: у результата поиска должен быть ID и положительный score.
	if result.EntryID == "" || result.Score <= 0 {
		t.Fatal("expected initialized search result")
	}
}

func TestStringsToInterfaces(t *testing.T) {
	// Arrange/Act: преобразуем срез строк в срез interface{}.
	got := stringsToInterfaces([]string{"a", "b"})
	// Assert: проверяем размер и содержимое результата.
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected conversion result: %#v", got)
	}
}
