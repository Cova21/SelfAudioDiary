package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHandler_RegisterInvalidJSON(t *testing.T) {
	// Arrange: создаём handler и запрос с невалидным JSON телом.
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	// Act: вызываем метод Register.
	h.Register(rr, req)

	// Assert: проверяем, что вернулся HTTP 400.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_LoginInvalidJSON(t *testing.T) {
	// Arrange: создаём handler и запрос с невалидным JSON для login.
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	// Act: вызываем метод Login.
	h.Login(rr, req)

	// Assert: ожидаем HTTP 400.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_LogoutMissingToken(t *testing.T) {
	// Arrange: формируем logout-запрос без Authorization заголовка.
	h := &AuthHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	// Act: вызываем Logout.
	h.Logout(rr, req)

	// Assert: ожидаем HTTP 401.
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestSearchHandler_UnauthorizedWithoutContext(t *testing.T) {
	// Arrange: создаём search-запрос без user_id в context.
	h := &SearchHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
	rr := httptest.NewRecorder()

	// Act: вызываем Search.
	h.Search(rr, req)

	// Assert: ожидаем HTTP 401.
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
