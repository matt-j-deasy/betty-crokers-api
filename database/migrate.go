package database

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

func RunMigrations(db *gorm.DB) error {
	slog.Info("Starting database migrations...")
	if err := db.AutoMigrate(
		&models.User{},
	); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	slog.Info("✅ GORM database migration completed successfully")
	return nil
}
