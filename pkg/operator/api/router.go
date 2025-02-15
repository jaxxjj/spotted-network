package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.ThrottleBacklog(50, 100, time.Second))

	// Task related endpoints
	r.Route("/api/v1", func(r chi.Router) {
		// Task operations
		r.Post("/tasks", h.SendRequest) // Create new task

		// Consensus queries
		r.Route("/consensus", func(r chi.Router) {
			r.Get("/", h.GetConsensusResponseByRequest)          // Query by parameters
			r.Get("/tasks/{taskID}", h.GetTaskConsensusByTaskID) // Query by task ID
		})
	})
}
