package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/rabbitmq"
	"go.uber.org/zap"
)

// NotificationType represents the type of notification
type NotificationType int

const (
	NotificationUnknown NotificationType = iota
	TranscriptionStarted
	TranscriptionCompleted
	TranscriptionFailed
	AnalysisStarted
	AnalysisCompleted
	AnalysisFailed
	EntryCreated
	EntryUpdated
	EntryDeleted
)

// Notification represents a notification message
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]string      `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	Read      bool                   `json:"read"`
}

// Client represents a WebSocket client
type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan *Notification
}

// NotificationService handles notification operations
type NotificationService struct {
	clients    map[string][]*Client
	clientsMux sync.RWMutex
	register   chan *Client
	unregister chan *Client
	mq         *rabbitmq.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(mq *rabbitmq.Client) *NotificationService {
	service := &NotificationService{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mq:         mq,
	}

	// Start the hub
	go service.run()

	// Start consuming notifications from RabbitMQ
	go service.consumeNotifications()

	return service
}

// run runs the notification hub
func (s *NotificationService) run() {
	for {
		select {
		case client := <-s.register:
			s.clientsMux.Lock()
			s.clients[client.UserID] = append(s.clients[client.UserID], client)
			s.clientsMux.Unlock()
			logger.Info("Client registered", zap.String("user_id", client.UserID))

		case client := <-s.unregister:
			s.clientsMux.Lock()
			if clients, ok := s.clients[client.UserID]; ok {
				// Remove client from the list
				for i, c := range clients {
					if c == client {
						s.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						close(client.Send)
						break
					}
				}
				// Remove user entry if no more clients
				if len(s.clients[client.UserID]) == 0 {
					delete(s.clients, client.UserID)
				}
			}
			s.clientsMux.Unlock()
			logger.Info("Client unregistered", zap.String("user_id", client.UserID))
		}
	}
}

// RegisterClient registers a new WebSocket client
func (s *NotificationService) RegisterClient(userID string, conn *websocket.Conn) *Client {
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan *Notification, 256),
	}
	s.register <- client
	return client
}

// UnregisterClient unregisters a WebSocket client
func (s *NotificationService) UnregisterClient(client *Client) {
	s.unregister <- client
}

// SendNotification sends a notification to a user
func (s *NotificationService) SendNotification(notification *Notification) error {
	s.clientsMux.RLock()
	clients, ok := s.clients[notification.UserID]
	s.clientsMux.RUnlock()

	if !ok || len(clients) == 0 {
		logger.Debug("No clients connected for user", zap.String("user_id", notification.UserID))
		return nil
	}

	// Send to all connected clients for this user
	for _, client := range clients {
		select {
		case client.Send <- notification:
		default:
			// Client's send channel is full, skip
			logger.Warn("Client send channel full", zap.String("user_id", notification.UserID))
		}
	}

	logger.Info("Notification sent",
		zap.String("user_id", notification.UserID),
		zap.String("type", fmt.Sprintf("%d", notification.Type)),
	)

	return nil
}

// PublishNotification publishes a notification to RabbitMQ
func (s *NotificationService) PublishNotification(notification *Notification) error {
	// Declare queue
	if err := s.mq.DeclareQueue("notification_queue"); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Publish message
	ctx := context.Background()
	if err := s.mq.Publish(ctx, "notification_queue", notification); err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	logger.Info("Notification published to queue",
		zap.String("user_id", notification.UserID),
		zap.String("type", fmt.Sprintf("%d", notification.Type)),
	)

	return nil
}

// consumeNotifications consumes notifications from RabbitMQ
func (s *NotificationService) consumeNotifications() {
	// Declare queue
	if err := s.mq.DeclareQueue("notification_queue"); err != nil {
		logger.Error("Failed to declare notification queue", zap.Error(err))
		return
	}

	// Start consuming
	handler := func(body []byte) error {
		var notification Notification
		if err := json.Unmarshal(body, &notification); err != nil {
			return fmt.Errorf("failed to unmarshal notification: %w", err)
		}

		return s.SendNotification(&notification)
	}

	if err := s.mq.Consume("notification_queue", handler); err != nil {
		logger.Error("Failed to consume notifications", zap.Error(err))
	}
}

// ReadClient reads messages from a WebSocket client
func (s *NotificationService) ReadClient(client *Client) {
	defer func() {
		s.UnregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}
	}
}

// WriteClient writes messages to a WebSocket client
func (s *NotificationService) WriteClient(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case notification, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(notification); err != nil {
				logger.Error("Failed to write message", zap.Error(err))
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
