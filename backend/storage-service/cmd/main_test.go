package main

import (
	"testing"

	"github.com/voice-diary/backend/internal/pkg/config"
)

func TestStorageMain_ConfigLoad(t *testing.T) {
	// Arrange: загружаем конфигурацию storage-service.
	cfg, err := config.Load()
	// Assert: загрузка конфигурации должна пройти успешно.
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	// Assert: ключевые поля MinIO должны быть заданы.
	if cfg.MinIOBucket == "" || cfg.MinIOEndpoint == "" {
		t.Fatal("expected non-empty MinIO config fields")
	}
}
