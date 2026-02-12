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
	"github.com/voice-diary/backend/api-gateway/handler"
	"github.com/voice-diary/backend/internal/pkg/config"
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

	logger.Info("Starting api-gateway", zap.String("environment", cfg.Environment))

	// Connect to services
	authConn, err := grpcutil.NewClient(cfg.AuthServiceAddr)
	if err != nil {
		logger.Fatal("Failed to connect to auth-service", zap.Error(err))
	}
	defer grpcutil.CloseConnection(authConn)

	diaryConn, err := grpcutil.NewClient(cfg.DiaryServiceAddr)
	if err != nil {
		logger.Fatal("Failed to connect to diary-service", zap.Error(err))
	}
	defer grpcutil.CloseConnection(diaryConn)

	storageConn, err := grpcutil.NewClient(cfg.StorageServiceAddr)
	if err != nil {
		logger.Fatal("Failed to connect to storage-service", zap.Error(err))
	}
	defer grpcutil.CloseConnection(storageConn)

	searchConn, err := grpcutil.NewClient(cfg.SearchServiceAddr)
	if err != nil {
		logger.Fatal("Failed to connect to search-service", zap.Error(err))
	}
	defer grpcutil.CloseConnection(searchConn)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authConn)
	diaryHandler := handler.NewDiaryHandler(diaryConn, storageConn)
	storageHandler := handler.NewStorageHandler(storageConn)
	searchHandler := handler.NewSearchHandler(searchConn)

	// Setup router
	router := setupRouter(authHandler, diaryHandler, storageHandler, searchHandler, cfg.JWTSecret)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func setupRouter(
	authHandler *handler.AuthHandler,
	diaryHandler *handler.DiaryHandler,
	storageHandler *handler.StorageHandler,
	searchHandler *handler.SearchHandler,
	jwtSecret string,
) http.Handler {
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Auth routes (no auth required)
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(jwtSecret))

	// Auth
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	// Diary entries
	api.HandleFunc("/entries", diaryHandler.CreateEntry).Methods("POST")
	api.HandleFunc("/entries", diaryHandler.ListEntries).Methods("GET")
	api.HandleFunc("/entries/upload", diaryHandler.UploadEntry).Methods("POST")
	api.HandleFunc("/entries/{id}", diaryHandler.GetEntry).Methods("GET")
	api.HandleFunc("/entries/{id}", diaryHandler.UpdateEntry).Methods("PUT")
	api.HandleFunc("/entries/{id}", diaryHandler.DeleteEntry).Methods("DELETE")

	// Storage
	api.HandleFunc("/storage/upload", storageHandler.GenerateUploadURL).Methods("POST")
	api.HandleFunc("/storage/{id}/download-url", storageHandler.GenerateDownloadURL).Methods("GET")
	api.HandleFunc("/storage/{id}/download", storageHandler.DownloadFile).Methods("GET")
	api.HandleFunc("/storage/{id}", storageHandler.DeleteFile).Methods("DELETE")

	// Search
	api.HandleFunc("/search", searchHandler.Search).Methods("GET")

	// Apply global middlewares
	handler := middleware.LoggingMiddleware(router)
	handler = middleware.SetupCORS().Handler(handler)

	return handler
}
