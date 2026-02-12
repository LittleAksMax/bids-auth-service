package api

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Role     string `json:"role" validate:"required,password"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
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

// InvalidateRefreshTokenRequest represents the request body for invalidating a refresh token.
type InvalidateRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
