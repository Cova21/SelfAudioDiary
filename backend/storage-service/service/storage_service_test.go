package service

import "testing"

func TestStorageServiceStructInitialization(t *testing.T) {
	// Arrange: создаём структуру сервиса с тестовым bucketName.
	s := &StorageService{bucketName: "voice-diary-audio"}
	// Assert: проверяем, что значение поля сохранилось корректно.
	if s.bucketName != "voice-diary-audio" {
		t.Fatalf("unexpected bucketName: %q", s.bucketName)
	}
}
