package api

import (
	"errors"
	"net/http"

	"github.com/LittleAksMax/bids-auth-service/internal/service"
)

// InvalidateRefreshToken invalidates a refresh token.
// Requires API key authentication via X-API-Key header (handled by middleware).
func (tc *TokensController) InvalidateRefreshToken(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[InvalidateRefreshTokenRequest](r)
	if body == nil {
		writeJSON(w, http.StatusBadRequest, Response{Success: false, Error: "invalid request body"})
		return
	}
	if err := tc.TokenMgr.InvalidateRefreshToken(r.Context(), body.RefreshToken); err != nil {
		if errors.Is(err, service.ErrInvalidRefresh) {
			writeJSON(w, http.StatusNotFound, Response{Success: false, Error: "refresh token not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to invalidate token"})
		return
	}
	writeJSON(w, http.StatusOK, Response{Success: true, Data: "token invalidated"})
}
