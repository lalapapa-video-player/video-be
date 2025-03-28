package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/godruoyi/go-snowflake"
	"github.com/lalapapa-video-player/video-be/internal/i"
	"github.com/lalapapa-video-player/video-be/internal/smbx"
	"github.com/sgostarter/i/commerr"
)

var (
	RootIDPrefixLocalDir = "L-"
	RootIDPrefixSMB      = "S-"
	RootIDPrefixHISTORY  = "H-"
)

type RootStat struct {
	ID   string `json:"id"` // prefix: L(local dir);S(smb); H(history)
	Name string `json:"name"`
}

type SMBRoot struct {
	Address  string
	User     string
	Password string
}

type TopRoots struct {
	RootStats []RootStat

	SMBRoots map[string]SMBRoot

	fsMap map[string]i.FS
}

func (tr *TopRoots) fix() {
	if len(tr.SMBRoots) == 0 {
		tr.SMBRoots = make(map[string]SMBRoot)
	}

	if len(tr.fsMap) == 0 {
		tr.fsMap = make(map[string]i.FS)
	}
}

type RootRequest struct {
	RType string `json:"rtype"`

	SMBAddress  string `json:"smb_address"`
	SMBUser     string `json:"smb_user"`
	SMBPassword string `json:"smb_password"`
}

type TestRootResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (s *Server) handleTestRoot(c *gin.Context) {
	var resp TestRootResponse

	err := s.handleTestRootInner(c)
	if err != nil {
		resp.StatusCode = -1
		resp.Message = err.Error()
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleTestRootInner(c *gin.Context) (err error) {
	var req RootRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	if !strings.Contains(req.SMBAddress, ":") {
		req.SMBAddress += ":445"
	}

	err = s.testRoot(&req)

	return
}

func (s *Server) testRoot(req *RootRequest) (err error) {
	if req.RType == "smb" {
		err = smbx.TestSmbConnect(req.SMBAddress, req.SMBUser, req.SMBPassword)
	} else if req.RType == "local" {
		err = commerr.ErrUnimplemented
	} else {
		err = commerr.ErrUnknown
	}

	return
}

func (s *Server) handleAddRoot(c *gin.Context) {
	id, err := s.handleAddRootInner(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (s *Server) handleAddRootInner(c *gin.Context) (id string, err error) {
	var req RootRequest

	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}

	err = s.testRoot(&req)
	if err != nil {
		return
	}

	if !strings.Contains(req.SMBAddress, ":") {
		req.SMBAddress += ":445"
	}

	err = s.roots.Change(func(o *TopRoots) (n *TopRoots, err error) {
		n = o
		if n == nil {
			n = &TopRoots{
				RootStats: make([]RootStat, 0, 2),
			}
		}

		n.fix()

		for _, root := range n.SMBRoots {
			if root.Address == req.SMBAddress && root.User == req.SMBUser {
				err = commerr.ErrExiting

				return
			}
		}

		id = fmt.Sprintf("%s%d", RootIDPrefixSMB, snowflake.ID())

		n.RootStats = append(n.RootStats, RootStat{
			ID:   id,
			Name: "SMB://" + req.SMBUser + "@" + req.SMBAddress,
		})

		n.SMBRoots[id] = SMBRoot{
			Address:  req.SMBAddress,
			User:     req.SMBUser,
			Password: req.SMBPassword,
		}

		n.fsMap[id] = smbx.NewSmbXProvider(req.SMBAddress, req.SMBUser, req.SMBPassword)

		return
	})

	return
}

func (s *Server) getFS(rid string) (fs i.FS) {
	s.roots.Read(func(d *TopRoots) {
		fs = d.fsMap[rid]
	})

	return
}

func (s *Server) getFSName(rid string) (name string) {
	s.roots.Read(func(d *TopRoots) {
		for _, stat := range d.RootStats {
			if stat.ID == rid {
				name = stat.Name

				break
			}
		}
	})

	return
}
