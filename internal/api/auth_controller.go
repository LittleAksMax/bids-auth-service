package api

import (
	"github.com/davidr/bids-auth-service/internal/service"
)

// AuthController houses dependencies for auth/token endpoints.
type AuthController struct {
	AuthService service.AuthService
}

// NewAuthController constructs an AuthController.
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		AuthService: authService,
	}
}
