package main

import (
	"testing"

	"github.com/voice-diary/backend/internal/pkg/config"
)

func TestDiaryMain_ConfigLoad(t *testing.T) {
	// Arrange: загружаем конфигурацию diary-service.
	cfg, err := config.Load()
	// Assert: загрузка должна быть успешной.
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	// Assert: ключевые поля конфигурации не должны быть пустыми.
	if cfg.RabbitMQHost == "" || cfg.DBHost == "" {
		t.Fatalf("unexpected empty required config fields: %+v", cfg)
	}
}
