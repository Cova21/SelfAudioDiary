package service

import "testing"

func TestNotificationTypeOrder(t *testing.T) {
	// Assert: enum-значения типов уведомлений должны идти в ожидаемом порядке.
	if !(NotificationUnknown < TranscriptionStarted &&
		TranscriptionStarted < TranscriptionCompleted &&
		TranscriptionCompleted < EntryDeleted) {
		t.Fatal("unexpected notification type order")
	}
}

func TestNotificationStructInitialization(t *testing.T) {
	// Arrange: создаём тестовую структуру уведомления.
	n := Notification{ID: "n1", UserID: "u1", Title: "title", Message: "msg"}
	// Assert: обязательные поля уведомления не должны быть пустыми.
	if n.ID == "" || n.UserID == "" {
		t.Fatal("notification required fields should not be empty")
	}
}
