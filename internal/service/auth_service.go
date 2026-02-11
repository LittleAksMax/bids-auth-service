package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LittleAksMax/bids-auth-service/internal/repository"
	"github.com/LittleAksMax/bids-auth-service/internal/token"
	"github.com/LittleAksMax/bids-util/passwords"
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
	credRepo repository.PasswordCredentialRepository
	tokenMgr *token.Manager
	pepper   string
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, credRepo repository.PasswordCredentialRepository, tokenMgr *token.Manager, pepper string) AuthService {
	return &authService{
		userRepo: userRepo,
		credRepo: credRepo,
		tokenMgr: tokenMgr,
		pepper:   pepper,
	}
}

// Register creates a new user account and returns tokens for immediate use.
func (s *authService) Register(ctx context.Context, username, email, password, role string) (string, *TokenPair, error) {
	// Normalise inputs
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))

	// Check uniqueness
	existingUser, _ := s.userRepo.FindByUsername(ctx, username)
	if existingUser != nil {
		return "", nil, ErrUserExists
	}
	existingEmail, _ := s.userRepo.FindByEmail(ctx, email)
	if existingEmail != nil {
		return "", nil, ErrUserExists
	}

	// NOTE: currently we are not using s.pepper
	hash, salt, err := passwords.HashPassword(password, passwords.DefaultParams)
	if err != nil {
		return "", nil, err
	}

	// Create user in repository
	userID, err := s.userRepo.Create(ctx, username, email, hash)
	if err != nil {
		return "", nil, ErrUserExists
	}
	// Store password credential
	err = s.credRepo.Create(ctx, userID, hash, salt)
	if err != nil {
		return "", nil, err
	}

	// Generate token pair for the new user
	refreshToken, accessToken, err := s.tokenMgr.GenerateTokenPair(ctx, userID, username, role)
	if err != nil {
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
	username = strings.TrimSpace(username)
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}
	cred, err := s.credRepo.GetByUserID(ctx, user.ID)
	if err != nil || cred == nil {
		return nil, ErrInvalidCredentials
	}
	valid, err := passwords.VerifyPassword(password, cred.PasswordHash, cred.PasswordSalt, passwords.DefaultParams)
	if err != nil || !valid {
		return nil, ErrInvalidCredentials
	}
	refreshToken, accessToken, err := s.tokenMgr.GenerateTokenPair(ctx, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
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

// Replace all password hashing and verification with passwords.Hash and passwords.Verify using DefaultParams and a pepper from config.
// Store both hash and salt in the PASSWORD_CREDENTIAL table.
// On registration, normalise email, check uniqueness, hash password with pepper+salt+params, store user and password credential.
// On login, fetch user by username/email, verify password with passwords.Verify.
// On refresh, rotate refresh token in DB as per new flow.
// On logout, revoke refresh token in DB.
