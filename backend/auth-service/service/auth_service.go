package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/voice-diary/backend/auth-service/domain"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	jwtSecret   string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

// Register registers a new user
func (s *AuthService) Register(email, password, username string) (*domain.User, string, string, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, "", "", fmt.Errorf("user with email %s already exists", email)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Username:     username,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &domain.Session{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Login authenticates a user
func (s *AuthService) Login(email, password string) (*domain.User, string, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID, user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &domain.Session{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Logout logs out a user
func (s *AuthService) Logout(accessToken string) error {
	return s.sessionRepo.Delete(accessToken)
}

// RefreshToken refreshes access and refresh tokens
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	// Get session by refresh token
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	// Delete old session
	if err := s.sessionRepo.Delete(session.AccessToken); err != nil {
		// Log error but continue
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create new session
	newSession := &domain.Session{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(newSession); err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// ValidateToken validates an access token
func (s *AuthService) ValidateToken(accessToken string) (string, string, error) {
	// Parse token
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing user_id in token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", "", fmt.Errorf("missing email in token")
	}

	// Check if session exists
	_, err = s.sessionRepo.GetByAccessToken(accessToken)
	if err != nil {
		return "", "", fmt.Errorf("session not found")
	}

	return userID, email, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = string(passwordHash)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Delete all sessions for the user
	if err := s.sessionRepo.DeleteByUserID(userID); err != nil {
		// Log error but don't fail
	}

	return nil
}

// generateTokens generates access and refresh tokens
func (s *AuthService) generateTokens(userID, email string) (string, string, error) {
	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"jti":     uuid.New().String(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}
