package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
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
	r.POST("/notes-content", s.SetNotesContent())
	r.GET("/notes-content", s.GetNote())
	r.GET("/notes-content/timestamp", s.GetTimestamp())

	r.Run(":" + s.Port)
}

const max = 1 << 17

func (s *Server) SetNotesContent() gin.HandlerFunc {
	maxHuman := humanize.Bytes(max)
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
			c.Status(http.StatusBadRequest)
			return
		}
		if len(body) > (max) {
			c.JSON(http.StatusBadRequest, "content too big Max:"+maxHuman)
			return
		}
		note := c.Query("note")
		if note == "" {
			c.JSON(http.StatusBadRequest, "missing note")
			return
		}
		con := s.Redis.Get()
		defer con.Close()

		now := time.Now().Format(time.RFC3339Nano)
		if _, err := con.Do("SET", note, body); err != nil {
			fmt.Println("Errors writing to redis", err)
			c.JSON(http.StatusBadRequest, "saving note")
			return
		}
		if _, err := con.Do("SET", "time:"+note, now); err != nil {
			fmt.Println("Errors writing to redis", err)
			c.JSON(http.StatusBadRequest, "saving note")
			return
		}
		c.JSON(200, now)
	}
}

func (s *Server) GetNote() gin.HandlerFunc {
	return func(c *gin.Context) {
		note := c.Query("note")
		if note == "" {
			c.JSON(http.StatusBadRequest, "missing note")
			return
		}

		if ts, err := s.getNoteRedis(note); err != nil {
			c.JSON(http.StatusInternalServerError, "note not found")
			return
		} else {
			if _, err := c.Writer.Write(ts); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
			return
		}
	}
}

func (s *Server) getNoteRedis(note string) ([]byte, error) {
	con := s.Redis.Get()
	defer con.Close()
	ts, err := redis.Bytes(con.Do("GET", note))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		fmt.Println("Errors reading from redis", err)
		return nil, err
	}
	return ts, nil
}

func (s *Server) GetTimestamp() gin.HandlerFunc {
	return func(c *gin.Context) {
		note := c.Query("note")
		if note == "" {
			c.JSON(http.StatusBadRequest, "missing note")
			return
		}
		if ts, err := s.getTimestampRedis(note); err != nil {
			c.Status(http.StatusInternalServerError)
		} else {
			c.JSON(200, ts)
		}
	}
}

func (s *Server) getTimestampRedis(note string) (string, error) {
	con := s.Redis.Get()
	defer con.Close()
	ts, err := redis.String(con.Do("GET", "time:"+note))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		}
		fmt.Println("Errors reading from redis", err)
		return "", err
	}
	return ts, nil
}

/* func (c *Client) readRedis(redisPool *redis.Pool) []byte {
	con := redisPool.Get()
	defer con.Close()
	reply, err := redis.Bytes(con.Do("GET", c.hub.key))
	if err != nil {
		if err != redis.ErrNil {
			fmt.Println(err)
		}
		return nil
	}
	return reply
}
func (c *Client) writeRedis(redisPool *redis.Pool, msg []byte) {
	con := redisPool.Get()
	defer con.Close()
} */

// Web socket vesion
// r.GET("/notes-content", s.NotesContent())
/* func (s *Server) NotesContent() gin.HandlerFunc {
	hubs := websocket.NewHubsMap()
	return func(c *gin.Context) {
		websocket.ServeWs(hubs, s.Redis, c.Writer, c.Request)
	}
} */

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
