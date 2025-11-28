package api

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LogoutRequest represents the request body for user logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshRequest represents the request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ValidateAccessTokenRequest represents the request body for validating an access token.
type ValidateAccessTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// InvalidateRefreshTokenRequest represents the request body for invalidating a refresh token.
type InvalidateRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
