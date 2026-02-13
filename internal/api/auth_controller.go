package api

import (
	"github.com/LittleAksMax/bids-auth-service/internal/service"
)

// AuthController houses dependencies for auth/token endpoints.
type AuthController struct {
	authService   service.AuthService
	tokenService  service.TokenService
	cookieService service.CookieService
}

// NewAuthController constructs an AuthController.
func NewAuthController(authService service.AuthService, tokenService service.TokenService, cookieService service.CookieService) *AuthController {
	return &AuthController{
		authService:   authService,
		tokenService:  tokenService,
		cookieService: cookieService,
	}
}
