package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/davidr/bids-auth-service/internal/db"
	"github.com/davidr/bids-auth-service/internal/repository"
	"github.com/davidr/bids-auth-service/internal/token"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username or email already exists")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrRefreshExpired     = errors.New("refresh token expired")
)

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// AuthService handles authentication business logic.
type AuthService interface {
	Register(ctx context.Context, username, email, password string) (userID string, tokens *TokenPair, err error)
	Login(ctx context.Context, username, password string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// authService implements AuthService.
type authService struct {
	userRepo repository.UserRepository
	tokenMgr *token.Manager
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, tokenMgr *token.Manager) AuthService {
	return &authService{
		userRepo: userRepo,
		tokenMgr: tokenMgr,
	}
}

// Register creates a new user account and returns tokens for immediate use.
func (s *authService) Register(ctx context.Context, username, email, password string) (string, *TokenPair, error) {
	// Normalize inputs
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))

	// Hash the password
	passwordHash, err := db.HashPassword(password)
	if err != nil {
		return "", nil, err
	}

	// Create user in repository
	userID, err := s.userRepo.Create(ctx, username, email, passwordHash)
	if err != nil {
		// Database constraint violations indicate duplicate username/email
		return "", nil, ErrUserExists
	}

	// Generate token pair for the new user
	refreshToken, accessToken, err := s.tokenMgr.GenerateTokenPair(ctx, userID)
	if err != nil {
		// Token generation failed - user was created but can't log in
		// This is logged but user can still login manually later
		return userID, nil, err
	}

	tokens := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return userID, tokens, nil
}

// Login authenticates a user and returns a token pair.
func (s *authService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	// Normalize username
	username = strings.TrimSpace(username)

	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !db.CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	refreshToken, accessToken, err := s.tokenMgr.GenerateTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout invalidates a refresh token.
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenMgr.InvalidateRefreshToken(ctx, refreshToken)
}

// RefreshTokens exchanges a refresh token for a new token pair.
func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate the old refresh token and get user ID
	userID, expiresAt, err := s.tokenMgr.GetRefreshTokenInfo(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		// Clean up expired token
		_ = s.tokenMgr.InvalidateRefreshToken(ctx, refreshToken)
		return nil, ErrRefreshExpired
	}

	// Invalidate old refresh token BEFORE generating new one
	// This prevents the old token from being reused if generation fails
	if err := s.tokenMgr.InvalidateRefreshToken(ctx, refreshToken); err != nil {
		// If we can't invalidate, don't proceed
		return nil, errors.New("failed to invalidate old token")
	}

	// Generate new token pair
	newRefreshToken, accessToken, err := s.tokenMgr.GenerateTokenPair(ctx, userID)
	if err != nil {
		// Token generation failed - old token is already invalidated
		// User will need to login again
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
