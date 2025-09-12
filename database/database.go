package database

import (
	"fmt"
	"net/url"

	"log/slog"

	"github.com/matt-j-deasy/betty-crokers-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var GormDB *gorm.DB

// InitializeDatabase sets up the GORM database connection
func InitializeDatabase(cfg config.Environment) (*gorm.DB, error) {
	encodedPassword := url.QueryEscape(cfg.DBPassword)

	sslMode := "disable"
	if cfg.RunMode == "prod" || cfg.RunMode == "staging" {
		sslMode = "require"
		slog.Info("Using SSL for database connection in production mode")
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, encodedPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, sslMode,
	)

	// Connect GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	GormDB = db
	slog.Info("✅ GORM connected to database")
	return db, nil
}

// CloseDatabase gracefully closes the database connection
func CloseDatabase() {
	sqlDB, err := GormDB.DB()
	if err != nil {
		slog.Error("Failed to get database connection", "error", err)
		return
	}
	sqlDB.Close()
	slog.Info("✅ Database connection closed")
}
