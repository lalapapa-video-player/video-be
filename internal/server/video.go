package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sgostarter/libeasygo/hash"
)

type VideoTmRequest struct {
	VideoID string `json:"vid"`
	Tm      int    `json:"tm"`
}

func (s *Server) handleVideoSaveTm(c *gin.Context) {
	err := s.handleVideoSaveTmInner(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (s *Server) getFileFromVideoID(videoID string) (file string, exists bool) {
	vi, ok := s.dCache.Get(s.videoIDKey(videoID))
	if !ok {
		return
	}

	file, ok = vi.(string)
	if !ok {
		return
	}

	exists = true

	return
}

func (s *Server) handleVideoSaveTmInner(c *gin.Context) (err error) {
	var req VideoTmRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	video, ok := s.getFileFromVideoID(req.VideoID)
	if !ok {
		err = errors.New("invalid video id")

		return
	}

	saveID := hash.MD5(video)

	err = s.lastVideoTms.Change(func(oldM map[string]LastVideoTmItem) (newM map[string]LastVideoTmItem, _ error) {
		newM = oldM

		if len(newM) == 0 {
			newM = make(map[string]LastVideoTmItem)
		}

		if req.Tm <= 0 {
			delete(newM, saveID)
		}

		newM[saveID] = LastVideoTmItem{
			Tm:       req.Tm,
			UpdateAt: time.Now(),
		}

		return
	})

	return
}

func (s *Server) getVideoTm(videoID string) (tm int) {
	video, ok := s.getFileFromVideoID(videoID)
	if !ok {
		return
	}

	saveID := hash.MD5(video)

	s.lastVideoTms.Read(func(m map[string]LastVideoTmItem) {
		tm = m[saveID].Tm
	})

	return
}
