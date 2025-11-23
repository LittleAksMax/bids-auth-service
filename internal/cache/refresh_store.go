package cache

import (
	"context"
	"time"
)

// RefreshTokenStore defines required operations for storing refresh tokens.
// Implementations should be concurrency-safe.
type RefreshTokenStore interface {
	Save(ctx context.Context, token string, userID string, expiresAt time.Time) error
	Get(ctx context.Context, token string) (userID string, expiresAt time.Time, err error)
	Delete(ctx context.Context, token string) error
}
