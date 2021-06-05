package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/jtorz/temp-secure-notes/app/config/serverconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Port         string
	AppMode      string
	LoggingLevel config.LogginLvl
	Redis        *redis.Pool
}

func NewServer() (*Server, error) {
	config, err := serverconfig.LoadConfig()
	if err != nil {
		return nil, err
	}
	server := Server{
		Port:         strconv.Itoa(config.Port),
		AppMode:      config.AppMode,
		LoggingLevel: config.LoggingLevel,
	}

	server.Redis, err = serverconfig.OpenRedis(config.RedisURL, 2, 2)
	if err != nil {
		return nil, fmt.Errorf("reddis open connection error: %w", err)
	}
	return &server, nil
}

func (s *Server) Start() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

	// Start server
	e.Logger.Fatal(e.Start(":" + s.Port))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
