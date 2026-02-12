package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/voice-diary/backend/notification-service/service"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, check the origin
		return true
	},
}

type WebSocketHandler struct {
	notificationService *service.NotificationService
}

func NewWebSocketHandler(notificationService *service.NotificationService) *WebSocketHandler {
	return &WebSocketHandler{
		notificationService: notificationService,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user_id from query params (in production, validate JWT token)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Register client
	client := h.notificationService.RegisterClient(userID, conn)

	// Start reading and writing
	go h.notificationService.WriteClient(client)
	h.notificationService.ReadClient(client)
}
