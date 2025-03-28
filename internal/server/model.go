package server

type PlayListPreviewRequest struct {
	Path           string `json:"path"`
	OnlyVideoFiles bool   `json:"only_video_files"`
}

type PlayListPreviewResponse struct {
	Items []string `json:"items"`
}

type PlayListSaveRequest struct {
	Path  string   `json:"path"`
	Items []string `json:"items"`
}

type VidePlayFinishedRequest struct {
	Path string `json:"path"`
}

type RemoveRootRequest struct {
	Path string `json:"path"`
}
