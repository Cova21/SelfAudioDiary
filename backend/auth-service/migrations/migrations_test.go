package migrations

import (
	"strings"
	"testing"
)

func TestGetMigrations_NotEmpty(t *testing.T) {
	// Arrange/Act: получаем список миграций.
	m := GetMigrations()
	// Assert: в списке должна быть хотя бы одна миграция.
	if len(m) == 0 {
		t.Fatal("expected at least one migration")
	}
	// Assert: первая миграция должна содержать SQL создания users.
	if !strings.Contains(m[0], "CREATE TABLE IF NOT EXISTS users") {
		t.Fatal("expected users table creation SQL")
	}
}