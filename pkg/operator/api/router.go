package api

import (
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	// create rate limiter: 2 requests per second, max burst 3 requests
	rateLimiter := NewRateLimiter(2, 3)
	
	// apply rate limit middleware
	r.Use(rateLimiter.RateLimit)
	
	// Task related endpoints
	r.Route("/api/v1", func(r chi.Router) {
		// Task operations
		r.Post("/tasks", h.SendRequest)                      // Create new task
		
		// Consensus queries
		r.Route("/consensus", func(r chi.Router) {
			r.Get("/", h.GetConsensusResponseByRequest)      // Query by parameters
			r.Get("/tasks/{taskID}", h.GetTaskConsensusByTaskID) // Query by task ID
		})
	})
} 
