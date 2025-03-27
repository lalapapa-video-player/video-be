package server

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cors "github.com/itsjamie/gin-cors"
	"github.com/lalapapa-video-player/video-be/internal/i"
	"github.com/lalapapa-video-player/video-be/internal/smbx"
	"github.com/patrickmn/go-cache"
)

//go:embed tpl/*.html
var tpl embed.FS

type Server struct {
	sMBs map[string]i.FS

	dCache *cache.Cache
}

func NewServer() *Server {
	s := &Server{
		sMBs:   make(map[string]i.FS),
		dCache: cache.New(time.Minute, time.Minute),
	}

	s.init()

	return s
}

func (s *Server) init() {
	s.sMBs["x"] = smbx.NewSmbXProvider("10.0.0.141:445", "x", "")

	go s.httpRoutine()
}

func (s *Server) Wait() {
	<-make(chan any)
}

func (s *Server) httpRoutine() {
	httpServer := gin.Default()

	httpServer.Use(cors.Middleware(cors.Config{
		ValidateHeaders: false,
		Origins:         "*",
		RequestHeaders:  "",
		ExposedHeaders:  "",
		Methods:         "",
		MaxAge:          0,
		Credentials:     false,
	}))

	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("secret"))
	httpServer.Use(sessions.Sessions("my-session", store))

	tmpl, err := template.ParseFS(tpl, "tpl/*.html")
	if err != nil {
		log.Fatal(err)
	}

	httpServer.SetHTMLTemplate(tmpl)

	httpServer.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "play.html", nil)
	})

	httpServer.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	httpServer.GET("/videos", s.ServeStreamFile)
	httpServer.POST("/browser", s.handleBrowser)

	httpServer.POST("/s-video-id", s.handleSVideoID)
	_ = httpServer.Run(":8200")
}

func (s *Server) ServeStreamFile(c *gin.Context) {
	err := s.serveStreamFile(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) serveStreamFile(c *gin.Context) (err error) {
	id := uuid.NewString()

	videoID, ok := c.GetQuery("file")
	if !ok {
		err = errors.New("no query")

		return
	}

	i, ok := s.dCache.Get(s.videoIDKey(videoID))
	if !ok {
		err = errors.New("no video id")

		return
	}

	video, ok := i.(string)
	if !ok {
		err = errors.New("invalid video id")

		return
	}

	fmt.Printf("%s before get stream\n", id)

	ss, err := s.GetStreamFile(video)
	if err != nil {
		fmt.Printf("%s before get stream: failed - %v\n", id, err)

		return
	}

	fmt.Printf("%s after get stream\n", id)

	defer ss.Close()

	fmt.Printf("%s before serve content\n", id)
	http.ServeContent(c.Writer, c.Request, video, ss.stat.ModTime(), ss.stream)
	fmt.Printf("%s after serve content\n", id)

	return
}
