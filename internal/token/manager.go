package token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/davidr/bids-auth-service/internal/cache"
)

// Claims represents access token payload.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Manager handles creation/validation of access and refresh tokens.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	store         cache.RefreshTokenStore
}

func NewManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, store cache.RefreshTokenStore) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		store:         store,
	}
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
func (m *Manager) ExchangeRefresh(ctx context.Context, refresh string) (string, error) {
	uid, exp, err := m.store.Get(ctx, refresh)
	if err != nil {
		return "", err
	}
	if time.Now().After(exp) {
		_ = m.store.Delete(ctx, refresh)
		return "", errors.New("refresh expired")
	}
	return m.newAccessToken(uid)
}

func (m *Manager) newAccessToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
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
