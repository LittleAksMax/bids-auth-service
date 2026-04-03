package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// contextKey type for context keys to avoid collisions.
type contextKey string

const requestBodyKey contextKey = "requestBody"

// RegisterMiddleware attaches common middleware to the router.
func RegisterMiddleware(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
}
