package domain

import (
	"time"
)

// User represents a user entity
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Username     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}

// Session represents a user session
type Session struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// SessionRepository defines the interface for session operations
type SessionRepository interface {
	Create(session *Session) error
	GetByAccessToken(token string) (*Session, error)
	GetByRefreshToken(token string) (*Session, error)
	Delete(accessToken string) error
	DeleteByUserID(userID string) error
}
