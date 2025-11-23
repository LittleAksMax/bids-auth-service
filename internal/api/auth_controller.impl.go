package api

import (
	"encoding/json"
	"net/http"
)

// apiResponse reused for controller JSON responses.
type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Health handler implementation.
func (c *AuthController) Health(w http.ResponseWriter, r *http.Request) {
	if err := c.DB.PingContext(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, apiResponse{Success: false, Error: "db not ready"})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "ok"})
}

// GenerateRefresh handler implementation.
func (c *AuthController) GenerateRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"user_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.UserID == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: "user_id required"})
		return
	}
	refresh, err := c.Mgr.GenerateRefreshToken(r.Context(), body.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: map[string]string{"refresh_token": refresh}})
}

// ExchangeRefresh handler implementation.
func (c *AuthController) ExchangeRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Refresh string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Refresh == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: "refresh_token required"})
		return
	}
	access, err := c.Mgr.ExchangeRefresh(r.Context(), body.Refresh)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: map[string]string{"access_token": access}})
}

// ValidateAccess handler implementation.
func (c *AuthController) ValidateAccess(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" || apiKey != c.Cfg.ValidationAPIKey {
		writeJSON(w, http.StatusForbidden, apiResponse{Success: false, Error: "invalid api key"})
		return
	}
	var body struct {
		Access string `json:"access_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Access == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Error: "access_token required"})
		return
	}
	claims, err := c.Mgr.ValidateAccess(body.Access)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, apiResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: claims})
}
