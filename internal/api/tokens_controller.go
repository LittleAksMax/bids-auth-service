package api

import (
	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/token"
)

// TokensController handles token management endpoints (validate, invalidate).
// These endpoints require API key authentication.
type TokensController struct {
	TokenMgr *token.Manager
	Cfg      *config.Config
}

// NewTokensController constructs a TokensController.
func NewTokensController(tokenMgr *token.Manager, cfg *config.Config) *TokensController {
	return &TokensController{
		TokenMgr: tokenMgr,
		Cfg:      cfg,
	}
}
