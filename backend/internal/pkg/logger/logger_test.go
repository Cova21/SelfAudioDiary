package logger

import "testing"

func TestGet_DefaultLoggerNotNil(t *testing.T) {
	// Act: получаем logger без предварительной инициализации.
	l := Get()
	// Assert: logger должен быть создан.
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestInit_WithInvalidLevelFallsBack(t *testing.T) {
	// Act: инициализируем logger с невалидным logLevel.
	if err := Init("test-service", "not-a-level", "development"); err != nil {
		// Assert: инициализация не должна завершаться ошибкой.
		t.Fatalf("Init() should not fail on invalid level: %v", err)
	}
}
