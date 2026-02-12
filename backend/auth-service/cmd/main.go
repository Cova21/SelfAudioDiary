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
	grpcHandler "github.com/voice-diary/backend/auth-service/delivery/grpc"
	"github.com/voice-diary/backend/auth-service/migrations"
	"github.com/voice-diary/backend/auth-service/repository/postgres"
	"github.com/voice-diary/backend/auth-service/repository/redis"
	"github.com/voice-diary/backend/auth-service/service"
	authpb "github.com/voice-diary/backend/internal/gen/auth"
	"github.com/voice-diary/backend/internal/pkg/config"
	"github.com/voice-diary/backend/internal/pkg/database"
	"github.com/voice-diary/backend/internal/pkg/grpcutil"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"github.com/voice-diary/backend/internal/pkg/middleware"
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

	logger.Info("Starting auth-service", zap.String("environment", cfg.Environment))

	// Connect to PostgreSQL
	db, err := database.ConnectPostgres(cfg.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(migrations.GetMigrations()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Connect to Redis
	redisClient, err := database.ConnectRedis(cfg.GetRedisAddr())
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.DB)
	sessionRepo := redis.NewSessionRepository(redisClient.Client)

	// Initialize service
	authService := service.NewAuthService(userRepo, sessionRepo, cfg.JWTSecret)

	// Initialize gRPC handler
	handler := grpcHandler.NewAuthHandler(authService)

	// Start gRPC server
	grpcServer := grpcutil.NewServer(cfg.GRPCPort)
	authpb.RegisterAuthServiceServer(grpcServer.Server, handler)

	go func() {
		logger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Start(); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Start HTTP health check server
	httpServer := setupHTTPServer(cfg.HTTPPort, db, redisClient)
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
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

func setupHTTPServer(port string, db *database.PostgresDB, redisClient *database.RedisClient) *http.Server {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database
		if err := db.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"database connection failed"}`))
			return
		}

		// Check Redis
		if err := redisClient.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"redis connection failed"}`))
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
