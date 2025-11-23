package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("refresh token not found")
	ErrExpired  = errors.New("refresh token expired")
)

type RedisRefreshStore struct {
	client *redis.Client
	keyNS  string // namespace prefix e.g. "refresh"
}

// NewRedisRefreshStore creates a Redis-backed RefreshTokenStore using values from Config.
// Returns an error if required Redis connection details are missing.
func NewRedisRefreshStore(cfg *config.Config) (*RedisRefreshStore, error) {
	if cfg.RedisHost == "" || cfg.RedisPort == "" {
		return nil, errors.New("redis host/port required in config for RedisRefreshStore")
	}
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &RedisRefreshStore{client: client, keyNS: "refresh"}, nil
}

func (s *RedisRefreshStore) buildKey(token string) string {
	return s.keyNS + ":" + token
}

// Save stores the refresh token with TTL derived from expiresAt.
func (s *RedisRefreshStore) Save(ctx context.Context, token string, userID string, expiresAt time.Time) error {
	if token == "" || userID == "" {
		return errors.New("token and userID required")
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("expiresAt must be in the future")
	}
	key := s.buildKey(token)
	// Store userID as value; expiration managed by Redis.
	return s.client.Set(ctx, key, userID, ttl).Err()
}

// Get retrieves the userID and calculates expiresAt using TTL.
func (s *RedisRefreshStore) Get(ctx context.Context, token string) (string, time.Time, error) {
	key := s.buildKey(token)
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", time.Time{}, ErrNotFound
		}
		return "", time.Time{}, err
	}
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return "", time.Time{}, err
	}
	if ttl <= 0 { // key exists but no TTL or expired
		return "", time.Time{}, ErrExpired
	}
	expiresAt := time.Now().Add(ttl)
	return val, expiresAt, nil
}

// Delete removes the refresh token key.
func (s *RedisRefreshStore) Delete(ctx context.Context, token string) error {
	key := s.buildKey(token)
	return s.client.Del(ctx, key).Err()
}

// Client exposes underlying redis client (optional future use).
func (s *RedisRefreshStore) Client() *redis.Client { return s.client }
