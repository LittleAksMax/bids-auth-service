package api

import (
	"errors"
	"net/http"

	"github.com/davidr/bids-auth-service/internal/service"
)

// ValidateAccessToken validates an access token and returns its claims.
// Requires API key authentication via X-Api-Key header (handled by middleware).
func (tc *TokensController) ValidateAccessToken(w http.ResponseWriter, r *http.Request) {

	body := GetRequestBody[ValidateAccessTokenRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Validate the access token
	claims, err := tc.TokenMgr.ValidateAccess(body.AccessToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid access token"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data: map[string]interface{}{
			"valid":      true,
			"user_id":    claims.UserID,
			"expires_at": claims.ExpiresAt.Time,
			"issued_at":  claims.IssuedAt.Time,
		},
	})
}

// InvalidateRefreshToken invalidates a refresh token.
// Requires API key authentication via X-API-Key header.
func (tc *TokensController) InvalidateRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Check API key
	apiKey := r.Header.Get(apiKeyHeader)
	if apiKey == "" || apiKey != tc.Cfg.ValidationAPIKey {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid or missing API key"})
		return
	}
	// Requires API key authentication via X-API-Key header (handled by middleware).
	body := GetRequestBody[InvalidateRefreshTokenRequest](r)
	if err := tc.TokenMgr.InvalidateRefreshToken(r.Context(), body.RefreshToken); err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, service.ErrInvalidRefresh) {
			writeJSON(w, http.StatusNotFound, apiResponse{Success: false, Error: "refresh token not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to invalidate token"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "token invalidated"})
}
