package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matt-j-deasy/betty-crokers-api/config"
)

// InitializeDatabase initializes a connection pool to the database
func InitializeDatabase(cfg config.Environment) (*pgxpool.Pool, error) {
	// Logging database connection attempt
	slog.Info("Connecting to database...")

	// Database connection configuration
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	dbConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		slog.Error("Failed to parse database URL", "error", err)
		return nil, err
	}

	// Set connection pool options
	dbConfig.MaxConns = 10 // Example value; adjust based on application needs
	dbConfig.MinConns = 2  // Keep some connections ready
	dbConfig.HealthCheckPeriod = 1 * time.Minute
	dbConfig.MaxConnLifetime = 1 * time.Hour
	dbConfig.MaxConnIdleTime = 30 * time.Minute

	// Try connecting to the database
	maxRetries := 5
	retryDelay := 2 * time.Second
	var pool *pgxpool.Pool

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err = pgxpool.NewWithConfig(ctx, dbConfig)
		if err == nil {
			// Verify connection with a health check
			if err = pool.Ping(ctx); err == nil {
				slog.Info("Successfully connected to the database.")
				return pool, nil
			}
		}

		slog.Warn("Unable to connect to database, retrying...", "attempt", i+1, "error", err)
		time.Sleep(retryDelay)
	}

	// All retries failed
	slog.Error("Max retry attempts reached. Unable to connect to the database.", "error", err)
	return nil, err
}
