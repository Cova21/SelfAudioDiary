package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	grpcHandler "github.com/voice-diary/backend/notification-service/delivery/grpc"
	wsHandler "github.com/voice-diary/backend/notification-service/delivery/websocket"
	"github.com/voice-diary/backend/notification-service/service"
	notificationpb "github.com/voice-diary/backend/internal/gen/notification"
	"github.com/voice-diary/backend/internal/pkg/config"
	"github.com/voice-diary/backend/internal/pkg/grpcutil"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/middleware"
	"github.com/voice-diary/backend/internal/pkg/rabbitmq"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.ServiceName, cfg.LogLevel, cfg.Environment); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting notification-service", zap.String("environment", cfg.Environment))

	// Connect to RabbitMQ
	mq, err := rabbitmq.NewClient(cfg.GetRabbitMQURL())
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer mq.Close()

	// Initialize service
	notificationService := service.NewNotificationService(mq)

	// Initialize handlers
	grpcNotificationHandler := grpcHandler.NewNotificationHandler(notificationService)
	wsNotificationHandler := wsHandler.NewWebSocketHandler(notificationService)

	// Start gRPC server
	grpcServer := grpcutil.NewServer(cfg.GRPCPort)
	notificationpb.RegisterNotificationServiceServer(grpcServer.Server, grpcNotificationHandler)

	go func() {
		logger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Start(); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Start HTTP/WebSocket server
	httpServer := setupHTTPServer(cfg.WSPort, wsNotificationHandler, mq)
	go func() {
		logger.Info("Starting WebSocket server", zap.String("port", cfg.WSPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start WebSocket server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	logger.Info("Servers stopped")
}

func setupHTTPServer(port string, wsHandler *wsHandler.WebSocketHandler, mq *rabbitmq.Client) *http.Server {
	router := mux.NewRouter()

	// WebSocket endpoint
	router.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check RabbitMQ
		if err := mq.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"rabbitmq connection failed"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Apply middlewares
	handler := middleware.LoggingMiddleware(router)
	handler = middleware.SetupCORS().Handler(handler)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
