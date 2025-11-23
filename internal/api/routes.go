package api

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all endpoint handlers using the controller methods.
func RegisterRoutes(r chi.Router, c *AuthController) {
	// Health
	r.Get("/health", c.Health)

	// Token routes
	r.Route("/tokens", func(r chi.Router) {
		// Generate refresh token
		r.Post("/refresh", c.GenerateRefresh)
		// Exchange refresh for access token
		r.Post("/exchange", c.ExchangeRefresh)
		// Validate access token (API key in header)
		r.Post("/validate", c.ValidateAccess)
	})
}
