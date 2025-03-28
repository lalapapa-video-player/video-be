package server

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cors "github.com/itsjamie/gin-cors"
	"github.com/lalapapa-video-player/video-be/internal/config"
	"github.com/lalapapa-video-player/video-be/internal/smbx"
	"github.com/patrickmn/go-cache"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

//go:embed tpl/*.html
var tpl embed.FS

//go:embed static/*
var fsStatic embed.FS

type Server struct {
	cfg *config.Config

	dCache *cache.Cache

	roots        *mwf.MemWithFile[*TopRoots, mwf.Serial, mwf.Lock]
	lastVideoTms *mwf.MemWithFile[map[string]LastVideoTmItem, mwf.Serial, mwf.Lock]
}

func (s *Server) BeforeLoad() {

}

func (s *Server) AfterLoad(r *TopRoots, err error) {
	if err != nil {
		return
	}

	r.fix()

	for id, root := range r.SMBRoots {
		r.fsMap[id] = smbx.NewSmbXProvider(root.Address, root.User, root.Password)
	}
}

func (s *Server) BeforeSave() {

}

func (s *Server) AfterSave(_ *TopRoots, _ error) {

}

type LastVideoTmItem struct {
	Tm       int       `json:"tm,omitempty"`
	UpdateAt time.Time `json:"ua,omitempty"`
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg:    cfg,
		dCache: cache.New(time.Minute, time.Minute),
		lastVideoTms: mwf.NewMemWithFileEx1[map[string]LastVideoTmItem, mwf.Serial, mwf.Lock](
			make(map[string]LastVideoTmItem), &mwf.JSONSerial{
				MarshalIndent: true,
			}, &sync.RWMutex{}, filepath.Join(cfg.DataRoot, "last_video_tms.json"), nil, nil, time.Second*5),
	}

	s.roots = mwf.NewMemWithFileEx[*TopRoots, mwf.Serial, mwf.Lock](
		&TopRoots{}, &mwf.JSONSerial{
			MarshalIndent: true,
		}, &sync.RWMutex{}, filepath.Join(cfg.DataRoot, "roots.json"), nil, s)

	s.init()

	return s
}

func (s *Server) init() {
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

	static, _ := fs.Sub(fsStatic, "static")
	httpServer.StaticFS("/static", http.FS(static))

	tmpl, err := template.ParseFS(tpl, "tpl/*.html")
	if err != nil {
		log.Fatal(err)
	}

	httpServer.SetHTMLTemplate(tmpl)

	httpServer.POST("/test-root", s.handleTestRoot)
	httpServer.POST("/add-root", s.handleAddRoot)

	httpServer.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	httpServer.POST("/browser", s.handleBrowser)

	httpServer.GET("/videos", s.ServeStreamFile)
	httpServer.POST("/video/save-tm", s.handleVideoSaveTm)

	httpServer.POST("/s-video-id", s.handleSVideoID)

	_ = httpServer.Run(s.cfg.Listen)
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

	video, ok := s.getFileFromVideoID(videoID)
	if !ok {
		err = errors.New("invalid video id")

		return
	}

	fmt.Printf("%s before get stream\n", id)

	ps := strings.SplitN(video, "/", 2)
	rID := ps[0]
	video = strings.Join(ps[1:], "/")

	ss, err := s.GetStreamFile(rID, video)
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
