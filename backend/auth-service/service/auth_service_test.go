package service

import (
	"errors"
	"testing"

	"github.com/voice-diary/backend/auth-service/domain"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createFn     func(user *domain.User) error
	getByIDFn    func(id string) (*domain.User, error)
	getByEmailFn func(email string) (*domain.User, error)
	updateFn     func(user *domain.User) error
}

func (m *mockUserRepo) Create(user *domain.User) error                      { return m.createFn(user) }
func (m *mockUserRepo) GetByID(id string) (*domain.User, error)             { return m.getByIDFn(id) }
func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error)       { return m.getByEmailFn(email) }
func (m *mockUserRepo) Update(user *domain.User) error                      { return m.updateFn(user) }
func (m *mockUserRepo) Delete(id string) error                              { return nil }

type mockSessionRepo struct {
	createFn           func(session *domain.Session) error
	getByAccessTokenFn func(token string) (*domain.Session, error)
	getByRefreshFn     func(token string) (*domain.Session, error)
	deleteFn           func(accessToken string) error
	deleteByUserIDFn   func(userID string) error
}

func (m *mockSessionRepo) Create(session *domain.Session) error {
	return m.createFn(session)
}
func (m *mockSessionRepo) GetByAccessToken(token string) (*domain.Session, error) {
	return m.getByAccessTokenFn(token)
}
func (m *mockSessionRepo) GetByRefreshToken(token string) (*domain.Session, error) {
	return m.getByRefreshFn(token)
}
func (m *mockSessionRepo) Delete(accessToken string) error { return m.deleteFn(accessToken) }
func (m *mockSessionRepo) DeleteByUserID(userID string) error {
	return m.deleteByUserIDFn(userID)
}

func TestRegisterSuccess(t *testing.T) {
	// Arrange: подготавливаем моки репозиториев и сервис авторизации.
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*domain.User, error) { return nil, errors.New("not found") },
		createFn: func(user *domain.User) error {
			user.ID = "u1"
			return nil
		},
		getByIDFn: func(id string) (*domain.User, error) { return nil, nil },
		updateFn:  func(user *domain.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{
		createFn: func(session *domain.Session) error { return nil },
		getByAccessTokenFn: func(token string) (*domain.Session, error) { return nil, nil },
		getByRefreshFn:     func(token string) (*domain.Session, error) { return nil, nil },
		deleteFn:           func(accessToken string) error { return nil },
		deleteByUserIDFn:   func(userID string) error { return nil },
	}
	svc := NewAuthService(userRepo, sessionRepo, "secret")

	// Act: вызываем Register с валидными данными.
	user, access, refresh, err := svc.Register("user@example.com", "password123", "alex")
	// Assert: проверяем отсутствие ошибки и корректные выходные данные.
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if user == nil || user.ID != "u1" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty access/refresh tokens")
	}
	if user.PasswordHash == "password123" {
		t.Fatal("password should be hashed")
	}
}

func TestRegisterExistingUserFails(t *testing.T) {
	// Arrange: мок userRepo возвращает существующего пользователя.
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*domain.User, error) { return &domain.User{ID: "exists"}, nil },
		createFn:     func(user *domain.User) error { return nil },
		getByIDFn:    func(id string) (*domain.User, error) { return nil, nil },
		updateFn:     func(user *domain.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{
		createFn: func(session *domain.Session) error { return nil },
		getByAccessTokenFn: func(token string) (*domain.Session, error) { return nil, nil },
		getByRefreshFn:     func(token string) (*domain.Session, error) { return nil, nil },
		deleteFn:           func(accessToken string) error { return nil },
		deleteByUserIDFn:   func(userID string) error { return nil },
	}
	svc := NewAuthService(userRepo, sessionRepo, "secret")

	// Act: пытаемся зарегистрировать пользователя с тем же email.
	_, _, _, err := svc.Register("user@example.com", "password123", "alex")
	// Assert: ожидаем ошибку.
	if err == nil {
		t.Fatal("expected error for existing user")
	}
}

func TestLoginSuccess(t *testing.T) {
	// Arrange: создаём валидный hash пароля и моки.
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: email, PasswordHash: string(hash)}, nil
		},
		createFn:  func(user *domain.User) error { return nil },
		getByIDFn: func(id string) (*domain.User, error) { return nil, nil },
		updateFn:  func(user *domain.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{
		createFn: func(session *domain.Session) error { return nil },
		getByAccessTokenFn: func(token string) (*domain.Session, error) { return nil, nil },
		getByRefreshFn:     func(token string) (*domain.Session, error) { return nil, nil },
		deleteFn:           func(accessToken string) error { return nil },
		deleteByUserIDFn:   func(userID string) error { return nil },
	}
	svc := NewAuthService(userRepo, sessionRepo, "secret")

	// Act: выполняем логин.
	user, access, refresh, err := svc.Login("user@example.com", "password123")
	// Assert: ожидаем успешный результат и непустые токены.
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}
	if user == nil || access == "" || refresh == "" {
		t.Fatal("expected user and tokens")
	}
}

func TestLoginInvalidPasswordFails(t *testing.T) {
	// Arrange: подготавливаем пользователя, но будем логиниться с неверным паролем.
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: email, PasswordHash: string(hash)}, nil
		},
		createFn:  func(user *domain.User) error { return nil },
		getByIDFn: func(id string) (*domain.User, error) { return nil, nil },
		updateFn:  func(user *domain.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{
		createFn: func(session *domain.Session) error { return nil },
		getByAccessTokenFn: func(token string) (*domain.Session, error) { return nil, nil },
		getByRefreshFn:     func(token string) (*domain.Session, error) { return nil, nil },
		deleteFn:           func(accessToken string) error { return nil },
		deleteByUserIDFn:   func(userID string) error { return nil },
	}
	svc := NewAuthService(userRepo, sessionRepo, "secret")

	// Act: вызываем Login с неправильным паролем.
	_, _, _, err := svc.Login("user@example.com", "wrong-password")
	// Assert: ожидаем ошибку.
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
}

func TestValidateTokenSuccess(t *testing.T) {
	// Arrange: настраиваем моки и получаем валидный access-токен.
	userRepo := &mockUserRepo{
		getByEmailFn: func(email string) (*domain.User, error) { return nil, nil },
		createFn:     func(user *domain.User) error { return nil },
		getByIDFn: func(id string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: "user@example.com"}, nil
		},
		updateFn: func(user *domain.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{
		createFn: func(session *domain.Session) error { return nil },
		getByAccessTokenFn: func(token string) (*domain.Session, error) {
			return &domain.Session{UserID: "u1", AccessToken: token}, nil
		},
		getByRefreshFn:   func(token string) (*domain.Session, error) { return nil, nil },
		deleteFn:         func(accessToken string) error { return nil },
		deleteByUserIDFn: func(userID string) error { return nil },
	}
	svc := NewAuthService(userRepo, sessionRepo, "secret")

	access, _, err := svc.generateTokens("u1", "user@example.com")
	if err != nil {
		t.Fatalf("generateTokens() error: %v", err)
	}
	// Act: валидируем токен через сервис.
	userID, email, err := svc.ValidateToken(access)
	// Assert: проверяем извлечённые claims.
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}
	if userID != "u1" || email != "user@example.com" {
		t.Fatalf("unexpected claims: userID=%s email=%s", userID, email)
	}
}
