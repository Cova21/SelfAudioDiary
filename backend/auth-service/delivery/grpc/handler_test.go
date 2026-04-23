package grpc

import (
	"testing"

	authpb "github.com/voice-diary/backend/internal/gen/auth"
)

func TestNewAuthHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём handler через конструктор.
	var svcServer authpb.AuthServiceServer = NewAuthHandler(nil)
	// Assert: результат должен быть не nil.
	if svcServer == nil {
		t.Fatal("expected non-nil auth gRPC server")
	}
}
