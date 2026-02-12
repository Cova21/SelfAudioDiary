package grpcutil

import (
	"context"
	"fmt"
	"net"

	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server wraps grpc.Server with additional functionality
type Server struct {
	*grpc.Server
	port string
}

// NewServer creates a new gRPC server with interceptors
func NewServer(port string) *Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor),
		grpc.StreamInterceptor(streamLoggingInterceptor),
	)

	return &Server{
		Server: server,
		port:   port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	logger.Info("gRPC server starting", zap.String("port", s.port))

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// GracefulStop stops the server gracefully
func (s *Server) GracefulStop() {
	logger.Info("Shutting down gRPC server gracefully")
	s.Server.GracefulStop()
}

// unaryLoggingInterceptor logs all unary RPC calls
func unaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	logger.Debug("gRPC unary call",
		zap.String("method", info.FullMethod),
	)

	resp, err := handler(ctx, req)

	if err != nil {
		logger.Error("gRPC unary call failed",
			zap.String("method", info.FullMethod),
			zap.Error(err),
		)
	}

	return resp, err
}

// streamLoggingInterceptor logs all streaming RPC calls
func streamLoggingInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	logger.Debug("gRPC stream call",
		zap.String("method", info.FullMethod),
	)

	err := handler(srv, ss)

	if err != nil {
		logger.Error("gRPC stream call failed",
			zap.String("method", info.FullMethod),
			zap.Error(err),
		)
	}

	return err
}

// GetUserIDFromContext extracts user_id from gRPC metadata
func GetUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing user_id in metadata")
	}

	return userIDs[0], nil
}
