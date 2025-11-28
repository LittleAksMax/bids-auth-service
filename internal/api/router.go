package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidr/bids-auth-service/internal/cache"
	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/health"
	"github.com/davidr/bids-auth-service/internal/repository"
	"github.com/davidr/bids-auth-service/internal/service"
	"github.com/davidr/bids-auth-service/internal/token"
)

// NewRouter constructs the main API router by wiring middleware and routes defined elsewhere.
func NewRouter(db *sql.DB, cfg *config.Config, store cache.RefreshTokenStore) http.Handler {
	r := chi.NewRouter()

	RegisterMiddleware(r)

	// Initialise layers
	userRepo := repository.NewUserRepository(db)
	tokenMgr := token.NewManager(cfg.AccessTokenSecret, cfg.RefreshTokenSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, store)
	authService := service.NewAuthService(userRepo, tokenMgr)

	// Initialise controllers
	authController := NewAuthController(authService)
	tokensController := NewTokensController(tokenMgr, cfg)

	// Create health checkers map
	healthCheckers := map[string]health.HealthChecker{
		"database": health.NewDBHealthChecker(db),
		"cache":    store, // RefreshTokenStore implements health.HealthChecker
	}

	RegisterRoutes(r, authController, tokensController, healthCheckers)

	return r
}
