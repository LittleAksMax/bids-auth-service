package cache

import (
	"context"
	"time"

	"github.com/davidr/bids-auth-service/internal/health"
)

// RefreshTokenStore defines required operations for storing refresh tokens.
// Implementations should be concurrency-safe and implement health.HealthChecker.
type RefreshTokenStore interface {
	health.HealthChecker
	Save(ctx context.Context, token string, userID string, expiresAt time.Time) error
	Get(ctx context.Context, token string) (userID string, expiresAt time.Time, err error)
	Delete(ctx context.Context, token string) error
}
