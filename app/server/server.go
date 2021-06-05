package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/jtorz/temp-secure-notes/app/config/serverconfig"
	"github.com/jtorz/temp-secure-notes/app/ctxinfo"
	"github.com/jtorz/temp-secure-notes/app/websocket"
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

	server.Redis, err = serverconfig.OpenRedis(config.RedisURL, config.RedisMaxOpen, config.RedisMaxIdle)
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
	r.Use(static.Serve("/", static.LocalFile("web/dist", true)))

	// Middleware used to add the app mode to the context.'
	r.Use(func(ginCtx *gin.Context) {
		ctxinfo.SetLoggingLevel(ginCtx, config.LogginLvl(s.LoggingLevel))
	})

	r.NoRoute(func(c *gin.Context) {
		defer c.Next()
		if strings.HasPrefix(c.Request.URL.Path, "/notes/") {
			c.Status(200)
			return
		}
	})

	r.LoadHTMLGlob("web/templates/*")
	r.Use(s.Notes())
	r.GET("/notes-content", s.NotesContent())
	r.Run(":" + s.Port)
}

func (s *Server) NotesContent() gin.HandlerFunc {
	hubs := websocket.NewHubsMap()
	return func(c *gin.Context) {
		websocket.ServeWs(hubs, s.Redis, c.Writer, c.Request)
	}
}

func (s *Server) Notes() gin.HandlerFunc {
	type TplData struct {
		NoteURL    string
		NoteURLEnc string
	}
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/notes/") {
			c.Next()
			return
		}
		key := c.Request.URL.String()
		c.HTML(http.StatusOK, "notes.html", TplData{
			NoteURL:    key,
			NoteURLEnc: url.QueryEscape(key),
		})
		c.Status(200)
		c.Abort()
	}
}
