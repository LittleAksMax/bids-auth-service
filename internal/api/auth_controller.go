package api

import (
	"database/sql"

	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/token"
)

// AuthController houses dependencies for auth/token endpoints.
type AuthController struct {
	DB  *sql.DB
	Mgr *token.Manager
	Cfg *config.Config
}

// NewAuthController constructs an AuthController.
func NewAuthController(db *sql.DB, mgr *token.Manager, cfg *config.Config) *AuthController {
	return &AuthController{DB: db, Mgr: mgr, Cfg: cfg}
}
