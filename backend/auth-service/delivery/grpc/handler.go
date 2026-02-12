package grpc

import (
	"context"

	"github.com/voice-diary/backend/auth-service/service"
	authpb "github.com/voice-diary/backend/internal/gen/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) authpb.AuthServiceServer {
	return &authHandler{
		authService: authService,
	}
}

func (h *authHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	user, accessToken, refreshToken, err := h.authService.Register(req.Email, req.Password, req.Username)
	if err != nil {
		return nil, err
	}

	return &authpb.RegisterResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours in seconds
	}, nil
}

func (h *authHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours in seconds
		User: &authpb.User{
			Id:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *authHandler) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	err := h.authService.Logout(req.AccessToken)
	if err != nil {
		return &authpb.LogoutResponse{Success: false}, err
	}

	return &authpb.LogoutResponse{Success: true}, nil
}

func (h *authHandler) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    86400, // 24 hours in seconds
	}, nil
}

func (h *authHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	userID, email, err := h.authService.ValidateToken(req.AccessToken)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Email:  email,
	}, nil
}

func (h *authHandler) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	err := h.authService.ChangePassword(req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		return &authpb.ChangePasswordResponse{Success: false}, err
	}

	return &authpb.ChangePasswordResponse{Success: true}, nil
}
