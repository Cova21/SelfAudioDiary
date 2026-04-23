package main

import (
	"testing"

	"github.com/voice-diary/backend/internal/pkg/config"
)

func TestNotificationMain_ConfigLoad(t *testing.T) {
	// Arrange: загружаем конфигурацию notification-service.
	cfg, err := config.Load()
	// Assert: ошибок быть не должно.
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}
	// Assert: адрес notification-service должен быть задан.
	if cfg.NotificationServiceAddr == "" {
		t.Fatal("expected non-empty notification service address")
	}
}
