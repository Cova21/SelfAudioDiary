package grpc

import (
	"testing"

	notificationpb "github.com/voice-diary/backend/internal/gen/notification"
)

func TestNewNotificationHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём gRPC handler уведомлений.
	var srv notificationpb.NotificationServiceServer = NewNotificationHandler(nil)
	// Assert: handler должен быть не nil.
	if srv == nil {
		t.Fatal("expected non-nil notification gRPC server")
	}
}
