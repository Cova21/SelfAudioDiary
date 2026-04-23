package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupRouter_HealthEndpoint(t *testing.T) {
	// Arrange: поднимаем роутер и формируем GET-запрос к /health.
	router := setupRouter(nil, nil, nil, nil, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Act: отправляем запрос в router.
	router.ServeHTTP(rr, req)

	// Assert: проверяем код ответа 200.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Assert: проверяем ожидаемое тело health-ответа.
	if body := rr.Body.String(); body != `{"status":"healthy"}` {
		t.Fatalf("unexpected health body: %q", body)
	}
}

func TestSetupRouter_HealthMethodNotAllowed(t *testing.T) {
	// Arrange: поднимаем роутер и формируем POST-запрос на /health.
	router := setupRouter(nil, nil, nil, nil, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	// Act: выполняем запрос.
	router.ServeHTTP(rr, req)

	// Assert: проверяем, что метод не разрешён (405).
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestSetupRouter_ProtectedRouteRequiresToken(t *testing.T) {
	// Arrange: поднимаем роутер и формируем запрос к защищённому /api роуту без токена.
	router := setupRouter(nil, nil, nil, nil, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rr := httptest.NewRecorder()

	// Act: отправляем запрос.
	router.ServeHTTP(rr, req)

	// Assert: проверяем, что возвращается 401 Unauthorized.
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
