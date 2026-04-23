package grpc

import (
	"testing"

	storagepb "github.com/voice-diary/backend/internal/gen/storage"
)

func TestNewStorageHandler_NotNil(t *testing.T) {
	// Arrange/Act: создаём gRPC handler storage-service.
	var srv storagepb.StorageServiceServer = NewStorageHandler(nil)
	// Assert: handler должен быть создан.
	if srv == nil {
		t.Fatal("expected non-nil storage gRPC server")
	}
}
