package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/voice-diary/backend/auth-service/domain"
)

type sessionRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(client *redis.Client) domain.SessionRepository {
	return &sessionRepository{
		client: client,
		ttl:    24 * time.Hour, // 24 hours
	}
}

func (r *sessionRepository) Create(session *domain.Session) error {
	ctx := context.Background()

	// Store session by access token
	accessKey := fmt.Sprintf("session:access:%s", session.AccessToken)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := r.client.Set(ctx, accessKey, sessionData, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session by access token: %w", err)
	}

	// Store session by refresh token
	refreshKey := fmt.Sprintf("session:refresh:%s", session.RefreshToken)
	if err := r.client.Set(ctx, refreshKey, sessionData, r.ttl*7).Err(); err != nil {
		return fmt.Errorf("failed to store session by refresh token: %w", err)
	}

	// Store user's active sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
	if err := r.client.SAdd(ctx, userSessionsKey, session.AccessToken).Err(); err != nil {
		return fmt.Errorf("failed to add to user sessions: %w", err)
	}

	return nil
}

func (r *sessionRepository) GetByAccessToken(token string) (*domain.Session, error) {
	ctx := context.Background()
	key := fmt.Sprintf("session:access:%s", token)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepository) GetByRefreshToken(token string) (*domain.Session, error) {
	ctx := context.Background()
	key := fmt.Sprintf("session:refresh:%s", token)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepository) Delete(accessToken string) error {
	ctx := context.Background()

	// Get session to find user ID
	session, err := r.GetByAccessToken(accessToken)
	if err != nil {
		return err
	}

	// Delete access token
	accessKey := fmt.Sprintf("session:access:%s", accessToken)
	if err := r.client.Del(ctx, accessKey).Err(); err != nil {
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	// Delete refresh token
	refreshKey := fmt.Sprintf("session:refresh:%s", session.RefreshToken)
	if err := r.client.Del(ctx, refreshKey).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	// Remove from user's sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", session.UserID)
	if err := r.client.SRem(ctx, userSessionsKey, accessToken).Err(); err != nil {
		return fmt.Errorf("failed to remove from user sessions: %w", err)
	}

	return nil
}

func (r *sessionRepository) DeleteByUserID(userID string) error {
	ctx := context.Background()
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)

	// Get all sessions for the user
	tokens, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete each session
	for _, token := range tokens {
		if err := r.Delete(token); err != nil {
			// Log error but continue
			continue
		}
	}

	// Delete the user sessions set
	if err := r.client.Del(ctx, userSessionsKey).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	return nil
}
