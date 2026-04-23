package main

import (
	"testing"

	"github.com/voice-diary/backend/internal/pkg/config"
)

func TestAuthMain_ConfigLoad(t *testing.T) {
	// Arrange: загружаем конфигурацию сервиса.
	cfg, err := config.Load()
	// Assert: загрузка должна быть успешной.
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	// Assert: проверяем наличие обязательных полей.
	if cfg.GRPCPort == "" || cfg.DBHost == "" {
		t.Fatalf("unexpected empty required config fields: %+v", cfg)
	}
}
