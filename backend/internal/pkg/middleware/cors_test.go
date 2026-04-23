package middleware

import "testing"

func TestSetupCORS_NotNil(t *testing.T) {
	// Arrange: вызываем фабрику CORS middleware.
	c := SetupCORS()
	// Assert: результат не должен быть nil.
	if c == nil {
		t.Fatal("expected non-nil CORS middleware")
	}
}
