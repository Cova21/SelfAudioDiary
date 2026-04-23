package main

import (
	"testing"

	"github.com/voice-diary/backend/internal/pkg/config"
)

func TestSearchMain_ConfigLoad(t *testing.T) {
	// Arrange: загружаем конфигурацию search-service.
	cfg, err := config.Load()
	// Assert: загрузка должна завершиться без ошибок.
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	// Assert: URL Elasticsearch должен быть задан.
	if cfg.ElasticsearchURL == "" {
		t.Fatal("expected non-empty elasticsearch URL")
	}
}
