package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-util/requests"
	"github.com/LittleAksMax/bids-util/validation"
	"github.com/go-chi/chi/v5"

	"github.com/LittleAksMax/bids-auth-service/internal/health"
)

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

		requests.WriteJSON(w, statusCode, requests.APIResponse{
			Success: true,
			Data:    statuses,
		})
	}
}

// RegisterRoutes registers all endpoint handlers using the controller methods.
func RegisterRoutes(r chi.Router, c *AuthController /*tc *TokensController,*/, healthCheckers map[string]health.HealthChecker) {
	// Health
	r.Get("/health", Health(healthCheckers))

	// Auth routes
	authValidationFuncs := []func(any) error{
		validation.ValidateRequiredFields,
		validation.ValidateUUIDs,
		validation.ValidateEmails,
		validation.ValidatePasswords,
		validation.ValidateRoles,
	}
	r.Route("/auth", func(r chi.Router) {
		r.With(requests.ValidateRequest[RegisterRequest](authValidationFuncs)).Post("/register", c.Register)
		r.With(requests.ValidateRequest[LoginRequest](authValidationFuncs)).Post("/login", c.Login)
		r.With(requests.ValidateRequest[LogoutRequest](authValidationFuncs)).Post("/logout", c.Logout)
		r.With(requests.ValidateRequest[RefreshRequest](authValidationFuncs)).Post("/refresh", c.Refresh)
	})
}
