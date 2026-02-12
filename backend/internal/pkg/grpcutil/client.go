package grpcutil

import (
	"context"
	"fmt"
	"time"

	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClient creates a new gRPC client connection
func NewClient(target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(clientLoggingInterceptor),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	logger.Info("gRPC client connected", zap.String("target", target))
	return conn, nil
}

// clientLoggingInterceptor logs all client RPC calls
func clientLoggingInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(start)

	if err != nil {
		logger.Error("gRPC client call failed",
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		logger.Debug("gRPC client call",
			zap.String("method", method),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// CloseConnection closes gRPC connection
func CloseConnection(conn *grpc.ClientConn) {
	if conn != nil {
		if err := conn.Close(); err != nil {
			logger.Error("Failed to close gRPC connection", zap.Error(err))
		}
	}
}
