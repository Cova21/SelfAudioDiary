package domain

import "testing"

func TestDomainEntitiesFields(t *testing.T) {
	// Arrange: создаём тестовые доменные сущности User и Session.
	u := User{ID: "u1", Email: "user@example.com", Username: "alex"}
	// Assert: обязательные поля User должны быть заполнены.
	if u.ID == "" || u.Email == "" || u.Username == "" {
		t.Fatal("user entity required fields should not be empty")
	}

	s := Session{UserID: "u1", AccessToken: "a", RefreshToken: "r"}
	// Assert: обязательные поля Session должны быть заполнены.
	if s.UserID == "" || s.AccessToken == "" || s.RefreshToken == "" {
		t.Fatal("session entity required fields should not be empty")
	}
}
