package api

import (
	"github.com/go-chi/chi/v5"
)

const (
	RateLimit = 10
	RateLimitBurst = 15
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	
	rateLimiter := NewRateLimiter(RateLimit, RateLimitBurst)
	
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
