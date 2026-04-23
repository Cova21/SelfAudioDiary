package postgres

import (
	"testing"

	"github.com/voice-diary/backend/auth-service/domain"
)

func TestNewUserRepository_NotNil(t *testing.T) {
	// Arrange/Act: создаём репозиторий через конструктор.
	repo := NewUserRepository(nil)
	// Assert: репозиторий должен быть создан.
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
	// Assert: проверяем соответствие интерфейсу доменного репозитория.
	var _ domain.UserRepository = repo
}
