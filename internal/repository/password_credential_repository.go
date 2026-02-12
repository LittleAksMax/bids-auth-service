package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/LittleAksMax/bids-auth-service/internal/contracts"
	"github.com/google/uuid"
)

type PasswordCredentialRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID uuid.UUID, hash, salt string) error
	GetByUserID(ctx context.Context, db *sql.DB, userID uuid.UUID) (*contracts.PasswordCredential, error)
}

type postgresPasswordCredentialRepository struct {
}

func NewPasswordCredentialRepository() PasswordCredentialRepository {
	return &postgresPasswordCredentialRepository{}
}

func (r *postgresPasswordCredentialRepository) Create(ctx context.Context, tx *sql.Tx, userID uuid.UUID, hash, salt string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO password_credential (user_id, password_hash, password_salt) VALUES ($1, $2, $3)`,
		userID, hash, salt,
	)
	return err
}

func (r *postgresPasswordCredentialRepository) GetByUserID(ctx context.Context, db *sql.DB, userID uuid.UUID) (*contracts.PasswordCredential, error) {
	cred := &contracts.PasswordCredential{}
	err := db.QueryRowContext(ctx,
		`SELECT user_id, password_hash, password_salt FROM password_credential WHERE user_id = $1`,
		userID,
	).Scan(&cred.UserID, &cred.PasswordHash, &cred.PasswordSalt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cred, nil
}
