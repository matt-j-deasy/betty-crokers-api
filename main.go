package main

import (
	"log/slog"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx with database/sql

	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/database"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
	"github.com/matt-j-deasy/betty-crokers-api/middleware"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
	"github.com/matt-j-deasy/betty-crokers-api/server"
	"github.com/matt-j-deasy/betty-crokers-api/services"
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

	// Initialize database (returns GORM DB)
	db, err := database.InitializeDatabase(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.CloseDatabase()

	// Run migrations
	err = database.RunMigrations(db)
	if err != nil {
		slog.Error("problem running migrations", "err", err)
		os.Exit(1)
	}

	// Initialize repositories
	repos, err := repositories.InitializeRepositories(db)
	if err != nil {
		slog.Error("Failed to initialize repositories", "err", err)
		os.Exit(1)
	}

	// Initialize services
	services, err := services.InitializeServices(repos, cfg)
	if err != nil {
		slog.Error("Failed to initialize services", "err", err)
		os.Exit(1)
	}

	// Initialize handlers
	handlers, err := handlers.InitializeHandlers(services, cfg)
	if err != nil {
		slog.Error("Failed to initialize handlers", "err", err)
		os.Exit(1)
	}

	// Create server
	s := server.CreateServer(cfg, db, handlers)

	addr := ":" + pickPort(os.Getenv("PORT"), cfg.LocalPort, 8080)

	// Start server
	switch cfg.RunMode {
	case "local":
		slog.Info("Starting server", "mode", "local", "addr", addr)
	case "production":
		slog.Info("Starting server", "mode", "production", "addr", addr)
	default:
		slog.Error("unknown RUN_MODE", "mode", cfg.RunMode)
		os.Exit(1)
	}

	// Start
	if err := s.Start(addr); err != nil {
		slog.Error("Server crashed", "err", err)
		os.Exit(1)
	}
}

// pickPort returns the first non-empty/valid port, as a string without a leading colon.
func pickPort(portEnv string, localPort int, fallback int) string {
	if portEnv != "" {
		// Basic sanity: ensure it's a valid integer > 0
		if p, err := strconv.Atoi(portEnv); err == nil && p > 0 {
			return strconv.Itoa(p)
		}
	}
	if localPort > 0 {
		return strconv.Itoa(localPort)
	}
	return strconv.Itoa(fallback)
}
