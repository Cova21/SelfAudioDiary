package grpc

import (
	"testing"

	searchpb "github.com/voice-diary/backend/internal/gen/search"
)

func TestNewSearchHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём gRPC handler search-service.
	var srv searchpb.SearchServiceServer = NewSearchHandler(nil)
	// Assert: handler должен быть не nil.
	if srv == nil {
		t.Fatal("expected non-nil search gRPC server")
	}
}
