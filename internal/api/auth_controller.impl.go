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
		return
	}

	// Set refresh token cookie (for browser clients)
	http.SetCookie(w, c.cookieService.CreateSetAuthCookie(tokenPair.RefreshToken))

	requests.WriteJSON(w, http.StatusCreated, requests.APIResponse{
		Success: true,
		Data: AuthResponseData{
			User: AuthUserResponse{
				ID:        user.ID.String(),
				Username:  user.Username,
				Email:     user.Email,
				UpdatedAt: user.UpdatedAt.String(),
				CreatedAt: user.CreatedAt.String(),
				Role:      user.Role,
			},
			Tokens: AuthTokensResponse{
				RefreshToken: tokenPair.RefreshToken,
				AccessToken:  tokenPair.AccessToken,
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
		return
	}

	// Set refresh token cookie (for browser clients)
	http.SetCookie(w, c.cookieService.CreateSetAuthCookie(tokenPair.RefreshToken))

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data: AuthResponseData{
			User: AuthUserResponse{
				ID:        user.ID.String(),
				Username:  user.Username,
				Email:     user.Email,
				UpdatedAt: user.UpdatedAt.String(),
				CreatedAt: user.CreatedAt.String(),
				Role:      user.Role,
			},
			Tokens: AuthTokensResponse{
				RefreshToken: tokenPair.RefreshToken,
				AccessToken:  tokenPair.AccessToken,
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
		Data: AuthTokensResponse{
			RefreshToken: newTokenPair.RefreshToken,
			AccessToken:  newTokenPair.AccessToken,
		},
	})
}
