package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/jtorz/temp-secure-notes/app/config/serverconfig"
	"github.com/jtorz/temp-secure-notes/app/ctxinfo"
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

	gin.SetMode(s.AppMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// gin.logger middleware added only on debug mode.
	if config.LogDebug >= s.LoggingLevel {
		r.Use(gin.Logger())
	}

	//static files
	r.Use(static.Serve("/", static.LocalFile("./web/dist", true)))

	// Middleware used to add the app mode to the context.'
	r.Use(func(ginCtx *gin.Context) {
		ctxinfo.SetLoggingLevel(ginCtx, config.LogginLvl(s.LoggingLevel))
	})

	r.Use(s.Notes())
	r.Run(":" + s.Port)
}

func (s *Server) Notes() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/notes") {
			c.Next()
			return
		}
		c.JSON(200, c.Request.URL.Path)
		c.Abort()
	}
}
