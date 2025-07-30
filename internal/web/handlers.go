package web

import (
	"charex/internal/extractors"
	"net/http"
)

type Server struct {
	hub              *Hub
	DataDir          string
	sakuraExtractor  extractors.Extractor
	janitorExtractor extractors.Extractor
}

func NewServer(hub *Hub, dataDir string, sakura, janitor extractors.Extractor) *Server {
	return &Server{
		hub:              hub,
		DataDir:          dataDir,
		sakuraExtractor:  sakura,
		janitorExtractor: janitor,
	}
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	serveWs(s, w, r)
}