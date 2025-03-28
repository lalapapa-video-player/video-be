package server

import (
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lalapapa-video-player/video-be/internal/utils"
	"github.com/sgostarter/i/commerr"
)

type BrowserRequest struct {
	Path string `json:"path"`
	Op   string `json:"op"`
	Dir  string `json:"dir"`
}

type BrowserResp struct {
	Path     string     `json:"path"`
	PathName string     `json:"pathName"`
	Items    []VOFSItem `json:"items"`
}

func (s *Server) handleBrowser(c *gin.Context) {
	curDir, curDirname, items, err := s.handleBrowserInner(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, BrowserResp{
		Path:     curDir,
		PathName: curDirname,
		Items:    items,
	})
}

func (s *Server) handleBrowserInner(c *gin.Context) (curDir, curDirname string, fsItems []VOFSItem, err error) {
	var req BrowserRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	curDir = strings.TrimPrefix(req.Path, "/")
TryTop:
	if curDir == "" && req.Dir == "" {
		s.roots.Read(func(d *TopRoots) {
			fsItems = make([]VOFSItem, 0, len(d.RootStats))

			for _, stat := range d.RootStats {
				fsItems = append(fsItems, VOFSItem{
					Name:  stat.Name,
					Path:  stat.ID,
					IsDir: true,
				})
			}
		})

		return
	}

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
			curDir = ""

			goto TryTop
		}
	}

	ps := strings.Split(curDir, "/")

	rFs := s.getFS(ps[0])
	if rFs == nil {
		err = commerr.ErrNotFound

		return
	}

	subDir := strings.Join(ps[1:], "/")
	curDirname = s.getFSName(ps[0]) + "/" + subDir

	fr, err := rFs.List(subDir)
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
			Path:  item.Path,
			Size:  fnSize(item.Stat),
			IsDir: item.Stat == nil || item.Stat.IsDir(),
		})
	}

	return
}
