package main

import (
	"log/slog"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx with database/sql

	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/database"
	"github.com/matt-j-deasy/betty-crokers-api/middleware"
	"github.com/matt-j-deasy/betty-crokers-api/server"
)

func main() {
	// Setup Logging
	slog.SetDefault(middleware.SetupLogger(os.Stdout))

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	// Initialize database
	dbPool, err := database.InitializeDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Run migrations
	err = database.RunMigrations(dbPool)
	if err != nil {
		slog.Error("problem running migrations", "err", err)
		os.Exit(1)
	}

	// Create server
	s := server.CreateServer(cfg, dbPool)

	// Start server
	if cfg.RunMode == "local" {
		slog.Info("Starting local execution")
		err = s.Start(":" + strconv.Itoa(cfg.LocalPort))
		if err != nil {
			slog.Error("Failed to start server", "err", err)
			os.Exit(1)
		}
	} else {
		slog.Info("unknown RUNMODE", "mode", cfg.RunMode)
		os.Exit(1)
	}
}
