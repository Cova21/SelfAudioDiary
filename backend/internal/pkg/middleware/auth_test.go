package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, secret string) string {
	t.Helper()

	claims := JWTClaims{
		UserID: "user-123",
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Arrange: создаём middleware и запрос без токена.
	h := AuthMiddleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Act: пропускаем запрос через middleware.
	h.ServeHTTP(rr, req)

	// Assert: проверяем, что вернулся 401.
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_BearerToken(t *testing.T) {
	// Arrange: готовим валидный токен и middleware.
	secret := "secret"
	token := signToken(t, secret)

	h := AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok || userID != "user-123" {
			t.Fatalf("unexpected user id: %q", userID)
		}
		email, ok := GetEmailFromContext(r.Context())
		if !ok || email != "user@example.com" {
			t.Fatalf("unexpected email: %q", email)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Act: выполняем запрос через middleware.
	h.ServeHTTP(rr, req)

	// Assert: статус должен быть 200, а claims проверяются внутри handler.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_QueryToken(t *testing.T) {
	// Arrange: формируем запрос с токеном в query-параметре.
	secret := "secret"
	token := signToken(t, secret)

	h := AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/audio?token="+token, nil)
	rr := httptest.NewRecorder()

	// Act: вызываем middleware.
	h.ServeHTTP(rr, req)

	// Assert: запрос должен пройти.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestContextHelpers(t *testing.T) {
	// Arrange: добавляем user_id и email в context.
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, "u1")
	ctx = context.WithValue(ctx, EmailKey, "u1@example.com")

	// Act + Assert: проверяем корректное извлечение user_id.
	if v, ok := GetUserIDFromContext(ctx); !ok || v != "u1" {
		t.Fatalf("unexpected user id from context: %q", v)
	}
	// Act + Assert: проверяем корректное извлечение email.
	if v, ok := GetEmailFromContext(ctx); !ok || v != "u1@example.com" {
		t.Fatalf("unexpected email from context: %q", v)
	}
}
