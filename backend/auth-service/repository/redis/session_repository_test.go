package redis

import (
	"testing"

	"github.com/voice-diary/backend/auth-service/domain"
)

func TestNewSessionRepository_NotNil(t *testing.T) {
	// Arrange/Act: создаём redis-репозиторий через конструктор.
	repo := NewSessionRepository(nil)
	// Assert: репозиторий должен быть создан.
	if repo == nil {
		t.Fatal("expected non-nil session repository")
	}
	// Assert: проверяем соответствие интерфейсу SessionRepository.
	var _ domain.SessionRepository = repo
}
