package config

import (
	"log/slog"

	"github.com/Netflix/go-env"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Environment struct {
	LocalPort   int    `env:"LOCAL_PORT" validate:"required"`
	DBHost      string `env:"DB_HOST" validate:"required"`
	DBPort      string `env:"DB_PORT" validate:"required"`
	DBUser      string `env:"DB_USER" validate:"required"`
	DBPassword  string `env:"DB_PASSWORD" validate:"required"`
	DBName      string `env:"DB_NAME" validate:"required"`
	RunMode     string `env:"RUN_MODE" validate:"required,oneof=local production"`
	FrontEndURL string `env:"FRONT_END_URL" validate:"required,url"`
}

func LoadConfig() (Environment, error) {
	var cfg Environment

	// Load .env
	err := godotenv.Load(".env")
	if err != nil {
		slog.Info("Error loading .env file")
	}

	_, err = env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		slog.Error("Problem reading environment config", "err", err)
		return cfg, err
	}

	// Validate configuration
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		slog.Error("Invalid configuration", "err", err)
		return cfg, err
	}

	// Additional Defaults
	if cfg.RunMode == "" {
		slog.Info("RUN_MODE not set, defaulting to local")
		cfg.RunMode = "local"
	}

	if cfg.LocalPort == 0 {
		slog.Info("LOCAL_PORT not set, defaulting to 8080")
		cfg.LocalPort = 8080
	}

	return cfg, nil
}
