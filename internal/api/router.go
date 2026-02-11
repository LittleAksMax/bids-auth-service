package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LittleAksMax/bids-auth-service/internal/config"
	"github.com/LittleAksMax/bids-auth-service/internal/health"
	"github.com/LittleAksMax/bids-auth-service/internal/repository"
	"github.com/LittleAksMax/bids-auth-service/internal/service"
	"github.com/LittleAksMax/bids-auth-service/internal/token"
)

// NewRouter constructs the main API router by wiring middleware and routes defined elsewhere.
func NewRouter(db *sql.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	RegisterMiddleware(r)

	// Initialise layers
	userRepo := repository.NewUserRepository(db)
	credRepo := repository.NewPasswordCredentialRepository(db)
	tokenMgr := token.NewManager(
		cfg.AccessTokenSecret,
		cfg.RefreshTokenSecret,
		cfg.TokenIssuer,
		cfg.TokenAudience,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL)
	authService := service.NewAuthService(userRepo, credRepo, tokenMgr, cfg.PasswordPepper)

	// Initialise controllers
	authController := NewAuthController(authService)
	tokensController := NewTokensController(tokenMgr, cfg)

	// Create health checkers map
	healthCheckers := map[string]health.HealthChecker{
		"database": health.NewDBHealthChecker(db),
	}

	RegisterRoutes(r, authController, tokensController, healthCheckers)

	return r
}
