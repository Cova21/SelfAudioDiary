package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewWebSocketHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём websocket handler.
	h := NewWebSocketHandler(nil)
	// Assert: handler должен быть создан.
	if h == nil {
		t.Fatal("expected non-nil websocket handler")
	}
}

func TestHandleWebSocket_MissingUserID(t *testing.T) {
	// Arrange: создаём websocket handler и запрос без user_id.
	h := NewWebSocketHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rr := httptest.NewRecorder()

	// Act: вызываем websocket endpoint.
	h.HandleWebSocket(rr, req)
	// Assert: ожидаем HTTP 400.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
