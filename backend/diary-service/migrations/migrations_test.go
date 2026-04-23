package migrations

import (
	"strings"
	"testing"
)

func TestGetMigrations_NotEmpty(t *testing.T) {
	// Arrange/Act: получаем список миграций diary-service.
	m := GetMigrations()
	// Assert: миграции должны существовать.
	if len(m) == 0 {
		t.Fatal("expected at least one migration")
	}
	// Assert: SQL должен содержать создание таблицы diary_entries.
	if !strings.Contains(m[0], "CREATE TABLE IF NOT EXISTS diary_entries") {
		t.Fatal("expected diary_entries table creation SQL")
	}
}
