package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/LittleAksMax/bids-auth-service/internal/contracts"
	"github.com/LittleAksMax/bids-auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService interface {
	CreateNewTokenPair(ctx context.Context, userID uuid.UUID, username, role string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type tokenService struct {
	pool             *sql.DB
	refreshTokenRepo repository.RefreshTokenRepository
	userRepo         repository.UserRepository
	accessSecret     []byte
	refreshSecret    []byte
	accessTTL        time.Duration
	refreshTTL       time.Duration
	issuer           string
	audience         string
}

func NewTokenService(pool *sql.DB, refreshTokenRepo repository.RefreshTokenRepository, userRepo repository.UserRepository, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, issuer, audience string) TokenService {
	return &tokenService{
		pool:             pool,
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		accessSecret:     []byte(accessSecret),
		refreshSecret:    []byte(refreshSecret),
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		issuer:           issuer,
		audience:         audience,
	}
}

// GenerateAccessToken creates a signed JWT with the specified claims.
func (s *tokenService) generateAccessToken(userID uuid.UUID, username, role string) (string, error) {
	now := time.Now()
	jti := uuid.New().String()

	claims := contracts.Claims{
		Role: role,
		Name: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// GenerateRefreshToken creates a cryptographically secure random token (32 bytes, base64url-encoded).
func (s *tokenService) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *tokenService) CreateNewTokenPair(ctx context.Context, userID uuid.UUID, username, role string) (string, string, error) { // Parse userID string to UUID
	// Transaction for atomic revocation + creation of tokens
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("couldn't rollback transaction: %v\n", err)
		}
	}()

	// Revoke all existing refresh tokens for this user to avoid multiple active sessions
	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, tx, userID); err != nil {
		return "", "", err
	}

	accessToken, err := s.generateAccessToken(userID, username, role)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Hash and store refresh token in database
	tokenHash := s.hashRefreshToken(refreshToken)
	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(s.refreshTTL)

	_, err = s.refreshTokenRepo.Create(ctx, tx, userID, tokenHash, issuedAt, expiresAt)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("couldn't commit transaction: %v\n", err)
	}

	return accessToken, refreshToken, nil
}

// Refresh validates a refresh token, rotates it, and returns a new access+refresh token pair.
func (s *tokenService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	if refreshToken == "" {
		return "", "", errors.New("missing refresh token")
	}

	// Hash provided token and look it up
	tokenHash := s.hashRefreshToken(refreshToken)
	existing, err := s.refreshTokenRepo.FindByHash(ctx, s.pool, tokenHash)
	if err != nil {
		return "", "", err
	}
	if existing == nil {
		return "", "", errors.New("invalid refresh token")
	}
	// Validate not revoked and not expired
	if existing.RevokedAt != nil {
		return "", "", errors.New("refresh token revoked")
	}
	if time.Now().After(existing.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	// Fetch user to populate claims
	user, err := s.userRepo.FindByID(ctx, s.pool, existing.UserID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found for refresh token")
	}

	// Begin rotation transaction
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("couldn't rollback transaction: %v\n", err)
		}
	}()

	// Create new refresh token
	newRefresh, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	newHash := s.hashRefreshToken(newRefresh)
	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(s.refreshTTL)
	newTokenID, err := s.refreshTokenRepo.Create(ctx, tx, existing.UserID, newHash, issuedAt, expiresAt)
	if err != nil {
		return "", "", err
	}

	// Revoke the old token with replacement tracking
	if err := s.refreshTokenRepo.RevokeWithReplacement(ctx, tx, existing.TokenID, newTokenID); err != nil {
		return "", "", err
	}

	// Generate new access token for user
	access, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("couldn't commit transaction: %v\n", err)
	}

	return access, newRefresh, nil
}

// Logout revokes a refresh token if present and active; idempotent otherwise.
func (s *tokenService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil // idempotent: nothing to do
	}
	hash := s.hashRefreshToken(refreshToken)
	existing, err := s.refreshTokenRepo.FindByHash(ctx, s.pool, hash)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil // not found: treat as success
	}
	if existing.RevokedAt != nil {
		return nil // already revoked: success
	}
	
	// Revoke single token (no replacement)
	return s.refreshTokenRepo.Revoke(ctx, s.pool, existing.TokenID)
}

// HashRefreshToken computes HMAC-SHA256 hash of the refresh token using the refresh secret.
func (s *tokenService) hashRefreshToken(token string) string {
	h := hmac.New(sha256.New, s.refreshSecret)
	h.Write([]byte(token))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
