package api

import (
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	// create rate limiter: 2 requests per second, max burst 3 requests
	rateLimiter := NewRateLimiter(2, 3)
	
	// apply rate limit middleware
	r.Use(rateLimiter.RateLimit)
	
	// register routes
	r.Post("/api/v1/task", h.SendRequest)
	r.Get("/api/v1/task/{taskID}/final", h.GetTaskFinalResponse)
} 