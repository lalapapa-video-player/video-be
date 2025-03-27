package main

import "github.com/lalapapa-video-player/video-be/internal/server"

func main() {
	s := server.NewServer()
	s.Wait()
}
