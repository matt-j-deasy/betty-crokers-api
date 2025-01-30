package database

import (
	"embed"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func RunMigrations(db *pgxpool.Pool) error {
	// Run migrations
	slog.Info("Running migrations...")
	goose.SetBaseFS(embeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("Goose dialogue failed", "err", err)
		return err
	}

	dbConn := stdlib.OpenDBFromPool(db)
	if err := goose.Up(dbConn, "migrations"); err != nil {
		slog.Error("Goose up failed", "err", err)
		return err
	}

	return nil
}
