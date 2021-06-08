package server

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	sConfig, err := serverconfig.LoadConfig()
	if err != nil {
		return nil, err
	}
	server := Server{
		Port:         strconv.Itoa(sConfig.Port),
		AppMode:      sConfig.AppMode,
		LoggingLevel: sConfig.LoggingLevel,
		MaxNoteSize:  uint32(sConfig.MaxNoteSize),
	}
	server.DataAccces, err = loadDataAccess(*sConfig)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func loadDataAccess(sConfig serverconfig.Config) (dataaccess.DataAccces, error) {
	if sConfig.InMemory {
		return dataaccess.NewInMemory(
			sConfig.DataTTL,
			int(sConfig.MaxNoteSize),
			int(sConfig.MaxNoteSize)*300, sConfig.LoggingLevel == config.LogDebug,
		)
	}
	redisPool, err := serverconfig.OpenRedis(sConfig.RedisURL, sConfig.RedisMaxOpen, sConfig.RedisMaxIdle)
	if err != nil {
		return nil, fmt.Errorf("reddis open connection error: %w", err)
	}
	return dataaccess.NewRedis(redisPool, sConfig.DataTTL), nil
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

	r.Use(func(c *gin.Context) {
		ctxinfo.SetLoggingLevel(c, s.LoggingLevel)
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=3600")
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
	r.GET("/goto", s.Goto())
	r.POST("/note", s.SetNoteContent())
	r.GET("/note", s.GetNote())
	r.GET("/note/v", s.GetNoteVersion())

	r.Run(":" + s.Port)
}

const missingNoteID = "uknown note"
const oopsError = "Oops!"

func (s *Server) Goto() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Query("note")
		if noteID == "" {
			url, err := getRand()
			if err != nil {
				s.logErr(c, err)
				c.Status(http.StatusBadRequest)
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, "/notes/"+url)
		}

		c.Redirect(http.StatusTemporaryRedirect, "/notes/"+noteID)
	}
}

func getRand() (s string, err error) {
	bytes := make([]byte, 16)
	if _, err = rand.Read(bytes); err != nil {
		return
	}
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(bytes), "/", "-"), nil
}

func (s *Server) SetNoteContent() gin.HandlerFunc {
	maxHuman := humanize.Bytes(uint64(s.MaxNoteSize))
	return func(c *gin.Context) {
		noteID := c.Query("n")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}

		note, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			s.logErr(c, err)
			c.Status(http.StatusBadRequest)
			return
		}
		if len(note) > (int(s.MaxNoteSize)) {
			c.JSON(http.StatusBadRequest, "content too big Max:"+maxHuman)
			return
		}

		version, err := s.DataAccces.SetNote(c, noteID, note)
		if err != nil {
			s.logErr(c, err)
			c.JSON(http.StatusInternalServerError, oopsError)
		}
		c.JSON(200, version)
	}
}

func (s *Server) GetNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Query("n")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}

		note, found, err := s.DataAccces.GetNote(c, noteID)
		if err != nil {
			s.logErr(c, err)
			c.JSON(http.StatusInternalServerError, oopsError)
			return
		} else if !found {
			c.Status(http.StatusOK)
			return
		}

		c.Header("Content-Type", "text/plain")
		c.Header("Content-Encoding", "gzip")
		zw := gzip.NewWriter(c.Writer)
		defer zw.Close()
		if _, err := zw.Write(note); err != nil {
			s.logErr(c, err)
			c.JSON(http.StatusInternalServerError, oopsError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func (s *Server) GetNoteVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		noteID := c.Query("n")
		if noteID == "" {
			c.JSON(http.StatusBadRequest, missingNoteID)
			return
		}
		if version, err := s.DataAccces.GetVersion(c, noteID); err != nil {
			s.logErr(c, err)
			c.JSON(http.StatusInternalServerError, oopsError)
		} else {
			c.JSON(200, version)
		}
	}
}

func (s *Server) Notes() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/notes/") {
			c.Next()
			return
		}
		if strings.HasSuffix(c.Request.URL.Path, "/") {
			c.Redirect(http.StatusMovedPermanently, c.Request.URL.Path[:len(c.Request.URL.Path)-1])
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "notes.html", nil)
		c.Status(200)
		c.Abort()
	}
}

func (s *Server) logErr(ctx context.Context, err error) {
	if ctxinfo.LogginAllowed(ctx, config.LogError) {
		log.Println(err)
	}
}
