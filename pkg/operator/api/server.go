package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	handler *Handler
	server  *http.Server
}

func NewServer(handler *Handler, port int) *Server {
	mux := http.NewServeMux()
	
	// Register routes
	mux.HandleFunc("/api/v1/task", handler.SendRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &Server{
		handler: handler,
		server:  server,
	}
}

func (s *Server) Start() error {
	log.Printf("Starting API server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
} 