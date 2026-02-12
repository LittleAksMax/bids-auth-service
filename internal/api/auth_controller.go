package api

import (
	"github.com/LittleAksMax/bids-auth-service/internal/service"
)

// AuthController houses dependencies for auth/token endpoints.
type AuthController struct {
	authService  service.AuthService
	tokenService service.TokenService
}

// NewAuthController constructs an AuthController.
func NewAuthController(authService service.AuthService, tokenService service.TokenService) *AuthController {
	return &AuthController{
		authService:  authService,
		tokenService: tokenService,
	}
}
