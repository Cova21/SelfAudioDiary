package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/voice-diary/backend/notification-service/service"
	notificationpb "github.com/voice-diary/backend/internal/gen/notification"
)

type notificationHandler struct {
	notificationpb.UnimplementedNotificationServiceServer
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) notificationpb.NotificationServiceServer {
	return &notificationHandler{
		notificationService: notificationService,
	}
}

func (h *notificationHandler) SendNotification(ctx context.Context, req *notificationpb.SendNotificationRequest) (*notificationpb.SendNotificationResponse, error) {
	notification := &service.Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserId,
		Type:      service.NotificationType(req.Type),
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		CreatedAt: time.Now(),
		Read:      false,
	}

	// Publish to RabbitMQ (will be consumed and sent to WebSocket clients)
	err := h.notificationService.PublishNotification(notification)
	if err != nil {
		return &notificationpb.SendNotificationResponse{
			Success:        false,
			NotificationId: "",
		}, err
	}

	return &notificationpb.SendNotificationResponse{
		Success:        true,
		NotificationId: notification.ID,
	}, nil
}

func (h *notificationHandler) Subscribe(req *notificationpb.SubscribeRequest, stream notificationpb.NotificationService_SubscribeServer) error {
	// This is a placeholder for gRPC streaming
	// The actual implementation uses WebSocket
	// This can be used for server-to-server communication if needed
	<-stream.Context().Done()
	return nil
}
