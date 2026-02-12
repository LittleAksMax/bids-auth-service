package api

import (
	"errors"
	"net/http"

	"github.com/LittleAksMax/bids-auth-service/internal/service"
)

// Register handler creates a new user account.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[RegisterRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	user, err := c.authService.Register(r.Context(), body.Username, body.Email, body.Password, body.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			writeJSON(w, http.StatusConflict, Response{Success: false, Error: "username or email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to register user"})
		return
	}

	accessToken, refreshToken, err := c.tokenService.CreateNewTokenPair(r.Context(), user.ID, user.Username, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to generate token pair"})
	}

	writeJSON(w, http.StatusCreated, Response{
		Success: true,
		Data: map[string]map[string]string{
			"user": {
				"id":         user.ID.String(),
				"username":   user.Username,
				"email":      user.Email,
				"created_at": user.CreatedAt.String(),
				"role":       user.Role,
			},
			"tokens": {
				"refresh_token": accessToken,
				"access_token":  refreshToken,
			},
		},
	})
}

// Login handler authenticates a user and returns both tokens.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[LoginRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to parse request"})
		return
	}

	// Obtain user and check password
	user, err := c.authService.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, Response{Success: false, Error: "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "login failed"})
		return
	}

	// Generate token pair
	accessToken, refreshToken, err := c.tokenService.CreateNewTokenPair(r.Context(), user.ID, user.Username, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to generate token pair"})
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"refresh_token": accessToken,
			"access_token":  refreshToken,
		},
	})
}

// Logout handler invalidates the refresh token.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[LogoutRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to parse request"})
		return
	}

	// Revoke the provided refresh token (idempotent)
	if err := c.tokenService.Logout(r.Context(), body.RefreshToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to logout"})
		return
	}

	// Clear refresh token cookie if present (Path must match where it's set, e.g., /auth/refresh)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth/refresh",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// Refresh handler exchanges a refresh token for a new token pair.
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	body := GetRequestBody[RefreshRequest](r)
	if body == nil {
		writeJSON(w, http.StatusInternalServerError, Response{Success: false, Error: "failed to parse request"})
		return
	}

	accessToken, newRefreshToken, err := c.tokenService.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, Response{Success: false, Error: "invalid or expired refresh token"})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]string{
			"refresh_token": newRefreshToken,
			"access_token":  accessToken,
		},
	})
}
