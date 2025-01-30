package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matt-j-deasy/betty-crokers-api/config"

	labmiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

type Server struct {
	*echo.Echo
	config config.Environment
}

func CreateServer(cfg config.Environment, dbPool *pgxpool.Pool) *Server {
	// Create server
	s := &Server{
		Echo:   echo.New(),
		config: cfg,
	}

	s.HideBanner = true

	s.Use(labmiddleware.CORSWithConfig(labmiddleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontEndURL},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
	}))

	// Public Routes
	s.GET("/health", HealthCheck)

	return s
}

func HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
