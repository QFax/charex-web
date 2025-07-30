package web

import (
	"net/http"
)

type Server struct {
	hub *Hub
}

func NewServer(hub *Hub) *Server {
	return &Server{hub: hub}
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	serveWs(s.hub, w, r)
}