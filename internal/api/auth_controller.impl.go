package api

import (
	"errors"
	"net/http"

	"github.com/LittleAksMax/bids-auth-service/internal/service"
	"github.com/LittleAksMax/bids-util/requests"
)

// Register handler creates a new user account.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	body := requests.GetRequestBody[RegisterRequest](r)
	if body == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Call service layer
	user, err := c.authService.Register(r.Context(), body.Username, body.Email, body.Password, body.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			requests.WriteJSON(w, http.StatusConflict, requests.APIResponse{Success: false, Error: "username or email already exists"})
			return
		}
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to register user"})
		return
	}

	tokenPair, err := c.tokenService.CreateNewTokenPair(r.Context(), user.ID, user.Username, user.Role)
	if err != nil || tokenPair == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to generate token pair"})
	}

	// Set refresh token cookie (for browser clients)
	http.SetCookie(w, c.cookieService.CreateSetAuthCookie(tokenPair.RefreshToken))

	requests.WriteJSON(w, http.StatusCreated, requests.APIResponse{
		Success: true,
		Data: map[string]map[string]string{
			"user": {
				"id":         user.ID.String(),
				"username":   user.Username,
				"email":      user.Email,
				"updated_at": user.UpdatedAt.String(),
				"created_at": user.CreatedAt.String(),
				"role":       user.Role,
			},
			"tokens": {
				"refresh_token": tokenPair.RefreshToken,
				"access_token":  tokenPair.AccessToken,
			},
		},
	})
}

// Login handler authenticates a user and returns both tokens.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	body := requests.GetRequestBody[LoginRequest](r)
	if body == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Obtain user and check password
	user, err := c.authService.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			requests.WriteJSON(w, http.StatusUnauthorized, requests.APIResponse{Success: false, Error: "invalid credentials"})
			return
		}
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "login failed"})
		return
	}

	// Generate token pair
	tokenPair, err := c.tokenService.CreateNewTokenPair(r.Context(), user.ID, user.Username, user.Role)
	if err != nil || tokenPair == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to generate token pair"})
	}

	// Set refresh token cookie (for browser clients)
	http.SetCookie(w, c.cookieService.CreateSetAuthCookie(tokenPair.RefreshToken))

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data: map[string]map[string]string{
			"user": {
				"id":         user.ID.String(),
				"username":   user.Username,
				"email":      user.Email,
				"updated_at": user.UpdatedAt.String(),
				"created_at": user.CreatedAt.String(),
				"role":       user.Role,
			},
			"tokens": {
				"refresh_token": tokenPair.RefreshToken,
				"access_token":  tokenPair.AccessToken,
			},
		},
	})
}

// Logout handler invalidates the refresh token.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	body := requests.GetRequestBody[LogoutRequest](r)
	if body == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to parse request"})
		return
	}

	// Revoke the provided refresh token (idempotent)
	if err := c.tokenService.Logout(r.Context(), body.RefreshToken); err != nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to logout"})
		return
	}

	// Clear refresh token cookie if present
	http.SetCookie(w, c.cookieService.CreateClearAuthCookie())

	w.WriteHeader(http.StatusNoContent)
}

// Refresh handler exchanges a refresh token for a new token pair.
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	body := requests.GetRequestBody[RefreshRequest](r)
	if body == nil {
		requests.WriteJSON(w, http.StatusInternalServerError, requests.APIResponse{Success: false, Error: "failed to parse request"})
		return
	}

	newTokenPair, err := c.tokenService.Refresh(r.Context(), body.RefreshToken)
	if err != nil || newTokenPair == nil {
		requests.WriteJSON(w, http.StatusUnauthorized, requests.APIResponse{Success: false, Error: "invalid or expired refresh token"})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data: map[string]string{
			"refresh_token": newTokenPair.RefreshToken,
			"access_token":  newTokenPair.AccessToken,
		},
	})
}
