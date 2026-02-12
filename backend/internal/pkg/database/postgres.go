package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// PostgresDB wraps sql.DB with additional functionality
type PostgresDB struct {
	*sql.DB
}

// ConnectPostgres creates a new PostgreSQL connection
func ConnectPostgres(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return &PostgresDB{db}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	logger.Info("Closing database connection")
	return db.DB.Close()
}

// HealthCheck checks if the database is healthy
func (db *PostgresDB) HealthCheck() error {
	ctx, cancel := getContext()
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// RunMigrations runs database migrations
func (db *PostgresDB) RunMigrations(migrations []string) error {
	logger.Info("Running database migrations", zap.Int("count", len(migrations)))

	for i, migration := range migrations {
		logger.Debug("Running migration", zap.Int("number", i+1))
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	logger.Info("All migrations completed successfully")
	return nil
}
