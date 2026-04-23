package grpc

import (
	"testing"

	diarypb "github.com/voice-diary/backend/internal/gen/diary"
)

func TestNewDiaryHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём gRPC handler diary-service.
	var srv diarypb.DiaryServiceServer = NewDiaryHandler(nil)
	// Assert: handler должен быть создан.
	if srv == nil {
		t.Fatal("expected non-nil diary gRPC server")
	}
}
