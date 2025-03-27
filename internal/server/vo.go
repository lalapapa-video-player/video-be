package server

type VOFSItem struct {
	Name  string `json:"name"`
	Size  string `json:"size"`
	IsDir bool   `json:"is_dir"`
}
