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
		statuses := make(HealthResponseData)
		allHealthy := true

		for name, checker := range checkers {
			if err := checker.HealthCheck(r.Context()); err != nil {
				statuses[name] = HealthServiceStatusResponse{
					Status: "unhealthy",
					Error:  err.Error(),
				}
				allHealthy = false
			} else {
				statuses[name] = HealthServiceStatusResponse{
					Status: "healthy",
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

	validationFuncs := []func(any) error{
		validation.ValidateRequiredFields,
		validation.ValidateUUIDs,
		validation.ValidatePasswords,
		validation.ValidateEmails,
		validation.ValidateRoles,
	}

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.With(requests.ValidateRequest[RegisterRequest](validationFuncs)).Post("/register", c.Register)
		r.With(requests.ValidateRequest[LoginRequest](validationFuncs)).Post("/login", c.Login)
		r.With(requests.ValidateRequest[LogoutRequest](validationFuncs)).Post("/logout", c.Logout)
		r.With(requests.ValidateRequest[RefreshRequest](validationFuncs)).Post("/refresh", c.Refresh)
	})
}
