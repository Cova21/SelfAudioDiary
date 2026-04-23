package database

import "testing"

func TestPostgresDBStructInitialization(t *testing.T) {
	// Arrange/Act: создаём пустую обёртку PostgresDB.
	db := &PostgresDB{}
	// Assert: указатель на структуру должен быть валидным.
	if db == nil {
		t.Fatal("expected non-nil PostgresDB pointer")
	}
}
