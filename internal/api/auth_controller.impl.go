package api

import (
	"errors"
	"net/http"

	"github.com/davidr/bids-auth-service/internal/service"
)

// apiResponse reused for controller JSON responses.
type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Register handler creates a new user account.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[RegisterRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	userID, tokens, err := c.AuthService.Register(r.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			writeJSON(w, http.StatusConflict, apiResponse{Success: false, Error: "username or email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to register user"})
		return
	}

	writeJSON(w, http.StatusCreated, apiResponse{
		Success: true,
		Data: map[string]string{
			"user_id":       userID,
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		},
	})
}

// Login handler authenticates a user and returns both tokens.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[LoginRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	tokens, err := c.AuthService.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "login failed"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data: map[string]string{
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		},
	})
}

// Logout handler invalidates the refresh token.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[LogoutRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	if err := c.AuthService.Logout(r.Context(), body.RefreshToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to logout"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "logged out"})
}

// Refresh handler exchanges a refresh token for a new token pair.
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[RefreshRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	tokens, err := c.AuthService.RefreshTokens(r.Context(), body.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefresh) || errors.Is(err, service.ErrRefreshExpired) {
			writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: "invalid or expired refresh token"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: "failed to refresh tokens"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data: map[string]string{
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		},
	})
}
