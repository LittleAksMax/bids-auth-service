package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidr/bids-auth-service/internal/health"
)

const apiKeyHeader = "X-Api-Key"

// Health handler implementation that checks all registered services.
// Always returns success=true since the request itself was fulfilled.
// Individual service health is reported in the data field.
func Health(checkers map[string]health.HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statuses := make(map[string]interface{})
		allHealthy := true

		for name, checker := range checkers {
			if err := checker.HealthCheck(r.Context()); err != nil {
				statuses[name] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				allHealthy = false
			} else {
				statuses[name] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}

		// Determine HTTP status code based on health
		statusCode := http.StatusOK
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		writeJSON(w, statusCode, apiResponse{
			Success: true,
			Data:    statuses,
		})
	}
}

// RegisterRoutes registers all endpoint handlers using the controller methods.
func RegisterRoutes(r chi.Router, c *AuthController, tc *TokensController, healthCheckers map[string]health.HealthChecker) {
	// Health
	r.Get("/health", Health(healthCheckers))

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.With(ValidateRequest[RegisterRequest]()).Post("/register", c.Register)
		r.With(ValidateRequest[LoginRequest]()).Post("/login", c.Login)
		r.With(ValidateRequest[LogoutRequest]()).Post("/logout", c.Logout)
		r.With(ValidateRequest[RefreshRequest]()).Post("/refresh", c.Refresh)
	})

	// Token management routes (API key protected)
	r.Route("/tokens", func(r chi.Router) {
		r.Use(RequireAPIKey(tc.Cfg.ValidationAPIKey))
		r.With(ValidateRequest[ValidateAccessTokenRequest]()).Post("/validate", tc.ValidateAccessToken)
		r.With(ValidateRequest[InvalidateRefreshTokenRequest]()).Post("/invalidate", tc.InvalidateRefreshToken)
	})
}
