package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the service
type Config struct {
	ServiceName string
	GRPCPort    string
	HTTPPort    string
	
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	
	// Redis
	RedisHost string
	RedisPort string
	
	// RabbitMQ
	RabbitMQHost string
	RabbitMQPort string
	RabbitMQUser string
	RabbitMQPass string
	
	// MinIO
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	
	// Elasticsearch
	ElasticsearchURL string
	
	// JWT
	JWTSecret string
	
	// Service addresses
	AuthServiceAddr         string
	DiaryServiceAddr        string
	StorageServiceAddr      string
	SearchServiceAddr       string
	NotificationServiceAddr string
	
	// WebSocket
	WSPort string
	
	// Environment
	Environment string
	LogLevel    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ServiceName: getEnv("SERVICE_NAME", "voice-diary-service"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		DBName:     getEnv("POSTGRES_DB", "voice_diary"),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		
		RabbitMQHost: getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort: getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUser: getEnv("RABBITMQ_USER", "guest"),
		RabbitMQPass: getEnv("RABBITMQ_PASS", "guest"),
		
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "voice-diary-audio"),
		MinIOUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		
		JWTSecret: getEnv("JWT_SECRET", "super-secret-jwt-key"),
		
		AuthServiceAddr:         getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		DiaryServiceAddr:        getEnv("DIARY_SERVICE_ADDR", "localhost:50052"),
		StorageServiceAddr:      getEnv("STORAGE_SERVICE_ADDR", "localhost:50053"),
		SearchServiceAddr:       getEnv("SEARCH_SERVICE_ADDR", "localhost:50054"),
		NotificationServiceAddr: getEnv("NOTIFICATION_SERVICE_ADDR", "localhost:50055"),
		
		WSPort: getEnv("WS_PORT", "8086"),
		
		Environment: getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
	
	return cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

// GetRedisAddr returns Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// GetRabbitMQURL returns RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		c.RabbitMQUser, c.RabbitMQPass, c.RabbitMQHost, c.RabbitMQPort,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}
