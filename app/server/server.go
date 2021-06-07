package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/jtorz/temp-secure-notes/app/config/serverconfig"
	"github.com/jtorz/temp-secure-notes/app/ctxinfo"
	"github.com/jtorz/temp-secure-notes/app/dataaccess"
)

type Server struct {
	Port         string
	AppMode      string
	MaxNoteSize  uint32
	LoggingLevel config.LogginLvl
	DataAccces   dataaccess.DataAccces
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
		MaxNoteSize:  1 << 17,
	}

	redisPool, err := serverconfig.OpenRedis(config.RedisURL, config.RedisMaxOpen, config.RedisMaxIdle)
	if err != nil {
		return nil, fmt.Errorf("reddis open connection error: %w", err)
	}
	server.DataAccces = dataaccess.NewRedis(redisPool, config.DataTTL)
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
	r.POST("/notes-content", s.SetNoteContent())
	r.GET("/notes-content", s.GetNote())
	r.GET("/notes-content/timestamp", s.GetNoteVersion())

	r.Run(":" + s.Port)
}

const missingNoteID = "unkown note"
const oopsError = "Oops!"

func (s *Server) SetNoteContent() gin.HandlerFunc {
	maxHuman := humanize.Bytes(uint64(s.MaxNoteSize))
	return func(c *gin.Context) {
		noteID := c.Query("note")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}

		note, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
			c.Status(http.StatusBadRequest)
			return
		}
		if len(note) > (int(s.MaxNoteSize)) {
			c.JSON(http.StatusBadRequest, "content too big Max:"+maxHuman)
			return
		}

		version, err := s.DataAccces.SetNote(c, noteID, note)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, oopsError)
		}
		c.JSON(200, version)
	}
}

func (s *Server) GetNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Query("note")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}

		note, found, err := s.DataAccces.GetNote(c, noteID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, oopsError)
			return
		} else if !found {
			c.Status(http.StatusOK)
			return
		}

		if _, err := c.Writer.Write(note); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, oopsError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func (s *Server) GetNoteVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Query("note")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}
		if version, err := s.DataAccces.GetVersion(c, noteID); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, oopsError)
		} else {
			c.JSON(200, version)
		}
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
		key := strings.TrimPrefix(c.Request.URL.Path, "/notes/")
		c.HTML(http.StatusOK, "notes.html", TplData{
			NoteURL:    key,
			NoteURLEnc: url.QueryEscape(key),
		})
		c.Status(200)
		c.Abort()
	}
}
