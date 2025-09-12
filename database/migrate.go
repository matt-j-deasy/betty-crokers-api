package database

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// RunMigrations applies database migrations using GORM.
func RunMigrations(db *gorm.DB) error {
	slog.Info("Starting database migrations...")

	err := db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	slog.Info("✅ GORM database migration completed successfully")
	return nil
}
