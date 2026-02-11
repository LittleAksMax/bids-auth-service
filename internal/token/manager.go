package token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents access token payload.
type Claims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles creation/validation of access and refresh tokens.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	audience      string
}

func NewManager(accessSecret, refreshSecret, issuer, audience string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        issuer,
		audience:      audience,
	}
}

// RefreshTokenStore defines the methods that a refresh token store must implement.
type RefreshTokenStore interface {
	Save(ctx context.Context, token, userID string, expires time.Time) error
	Get(ctx context.Context, token string) (userID string, expiresAt time.Time, err error)
	Delete(ctx context.Context, token string) error
}

// GenerateRefreshToken issues a random refresh token and stores it.
func (m *Manager) GenerateRefreshToken(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("userID required")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	tok := base64.RawURLEncoding.EncodeToString(b)
	expires := time.Now().Add(m.refreshTTL)
	if err := m.store.Save(ctx, tok, userID, expires); err != nil {
		return "", err
	}
	return tok, nil
}

// ExchangeRefresh validates the refresh token then returns a new signed access token.
// You must provide name and role for the user.
func (m *Manager) ExchangeRefresh(ctx context.Context, refresh, name, role string) (string, error) {
	uid, exp, err := m.store.Get(ctx, refresh)
	if err != nil {
		return "", err
	}
	if time.Now().After(exp) {
		_ = m.store.Delete(ctx, refresh)
		return "", errors.New("refresh expired")
	}
	return m.newAccessToken(uid, name, role)
}

func (m *Manager) newAccessToken(userID, name, role string) (string, error) {
	claims := Claims{
		Name: name,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID,
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.accessSecret)
}

// ValidateAccess verifies an access token signature and expiration, returning claims.
func (m *Manager) ValidateAccess(access string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(access, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GenerateTokenPair creates both a refresh token and access token for a user.
func (m *Manager) GenerateTokenPair(ctx context.Context, userID, name, role string) (refreshToken, accessToken string, err error) {
	refreshToken, err = m.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return "", "", err
	}
	accessToken, err = m.newAccessToken(userID, name, role)
	if err != nil {
		_ = m.store.Delete(ctx, refreshToken)
		return "", "", err
	}
	return refreshToken, accessToken, nil
}

// InvalidateRefreshToken removes a refresh token from the store.
func (m *Manager) InvalidateRefreshToken(ctx context.Context, refreshToken string) error {
	return m.store.Delete(ctx, refreshToken)
}

// GetRefreshTokenInfo retrieves the userID and expiration for a refresh token.
func (m *Manager) GetRefreshTokenInfo(ctx context.Context, refreshToken string) (userID string, expiresAt time.Time, err error) {
	return m.store.Get(ctx, refreshToken)
}
