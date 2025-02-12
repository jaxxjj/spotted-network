package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	handler *Handler
	server  *http.Server
}

func NewServer(handler *Handler, port int) *Server {
	r := chi.NewRouter()

	// Register routes
	handler.RegisterRoutes(r)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
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
