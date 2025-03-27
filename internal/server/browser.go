package server

import (
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lalapapa-video-player/video-be/internal/utils"
)

type BrowserRequest struct {
	Root string `json:"root"`
	Op   string `json:"op"`
	Dir  string `json:"dir"`
}

type BrowserResp struct {
	Root  string     `json:"root"`
	Items []VOFSItem `json:"items"`
}

func (s *Server) handleBrowser(c *gin.Context) {
	curDir, items, err := s.handleBrowserInner(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, BrowserResp{
		Root:  curDir,
		Items: items,
	})
}

func (s *Server) handleBrowserInner(c *gin.Context) (curDir string, fsItems []VOFSItem, err error) {
	var req BrowserRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	curDir = strings.TrimPrefix(req.Root, "/")

	if req.Op == "enter" {
		if curDir == "" {
			curDir = req.Dir
		} else {
			curDir = filepath.Join(curDir, req.Dir)
		}
	} else if req.Op == "leave" {
		if curDir == "" {
			err = errors.New("top")

			return
		}

		curDir = strings.TrimSuffix(curDir, "/")

		index := strings.LastIndex(curDir, "/")

		if index != -1 {
			curDir = curDir[0:index]
		} else {
			curDir = "/"
		}
	}

	fr, err := s.sMBs["x"].List(curDir)
	if err != nil {
		return
	}

	fsItems = make([]VOFSItem, 0, len(fr))

	fnSize := func(stat fs.FileInfo) string {
		if stat == nil {
			return ""
		}

		return utils.FormatSizePrecise(stat.Size())
	}

	for _, item := range fr {
		fsItems = append(fsItems, VOFSItem{
			Name:  item.Path,
			Size:  fnSize(item.Stat),
			IsDir: item.Stat == nil || item.Stat.IsDir(),
		})
	}

	return
}
