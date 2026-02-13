package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/LittleAksMax/bids-auth-service/internal/contracts"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	// Create stores a new refresh token in the database.
	Create(ctx context.Context, tx *sql.Tx, userID uuid.UUID, tokenHash string, issuedAt time.Time, expiresAt time.Time) (uuid.UUID, error)

	// FindByHash retrieves a refresh token by its hash.
	FindByHash(ctx context.Context, db *sql.DB, tokenHash string) (*contracts.RefreshToken, error)

	// Revoke marks a refresh token as revoked.
	Revoke(ctx context.Context, db *sql.DB, tokenID uuid.UUID) error

	// RevokeWithReplacement marks a token as revoked and records its replacement (for token rotation).
	RevokeWithReplacement(ctx context.Context, tx *sql.Tx, tokenID, replacementTokenID uuid.UUID) error

	// RevokeAllForUser revokes all active refresh tokens for a user (for logout all devices).
	RevokeAllForUser(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error

	// DeleteExpired removes expired tokens from the database (for cleanup).
	DeleteExpired(ctx context.Context, db *sql.DB) error
}

type refreshTokenRepository struct {
}

func NewRefreshTokenRepository() RefreshTokenRepository {
	return &refreshTokenRepository{}
}

// Create stores a new refresh token in the database.
func (r *refreshTokenRepository) Create(ctx context.Context, tx *sql.Tx, userID uuid.UUID, tokenHash string, issuedAt time.Time, expiresAt time.Time) (uuid.UUID, error) {
	tokenID := uuid.New()
	_, err := tx.ExecContext(ctx,
		`INSERT INTO refresh_tokens (token_id, user_id, token_hash, issued_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		tokenID, userID, tokenHash, issuedAt, expiresAt)
	if err != nil {
		return uuid.Nil, err
	}
	return tokenID, nil
}

// FindByHash retrieves a refresh token by its hash.
func (r *refreshTokenRepository) FindByHash(ctx context.Context, db *sql.DB, tokenHash string) (*contracts.RefreshToken, error) {
	query := `
		SELECT token_id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by_token_id
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var rt contracts.RefreshToken
	err := db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.TokenID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.IssuedAt,
		&rt.ExpiresAt,
		&rt.RevokedAt,
		&rt.ReplacedByTokenID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Revoke marks a refresh token as revoked.
func (r *refreshTokenRepository) Revoke(ctx context.Context, db *sql.DB, tokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_id = $1 AND revoked_at IS NULL
	`
	_, err := db.ExecContext(ctx, query, tokenID)
	return err
}

// RevokeWithReplacement marks a token as revoked and records its replacement (for token rotation).
func (r *refreshTokenRepository) RevokeWithReplacement(ctx context.Context, tx *sql.Tx, tokenID, replacementTokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), replaced_by_token_id = $2
		WHERE token_id = $1 AND revoked_at IS NULL
	`
	_, err := tx.ExecContext(ctx, query, tokenID, replacementTokenID)
	return err
}

// RevokeAllForUser revokes all active refresh tokens for a user (for logout all devices).
func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := tx.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired removes expired tokens from the database (for cleanup).
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context, db *sql.DB) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`
	_, err := db.ExecContext(ctx, query)
	return err
}
