package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidr/bids-auth-service/internal/cache"
	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/token"
)

// NewRouter constructs the main API router by wiring middleware and routes defined elsewhere.
func NewRouter(db *sql.DB, cfg *config.Config, store cache.RefreshTokenStore) http.Handler {
	r := chi.NewRouter()

	RegisterMiddleware(r)

	mgr := token.NewManager(cfg.AccessTokenSecret, cfg.RefreshTokenSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, store)
	controller := NewAuthController(db, mgr, cfg)
	RegisterRoutes(r, controller)

	return r
}
