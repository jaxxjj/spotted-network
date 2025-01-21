package api

import (
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/task", h.SendRequest)
	r.Get("/api/v1/task/{taskID}/final", h.GetTaskFinalResponse)
} 