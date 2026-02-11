package repository

import (
	"context"
	"database/sql"
)

type PasswordCredential struct {
	UserID       string
	PasswordHash string
	PasswordSalt string
}

type PasswordCredentialRepository interface {
	Create(ctx context.Context, userID, hash, salt string) error
	GetByUserID(ctx context.Context, userID string) (*PasswordCredential, error)
}

type postgresPasswordCredentialRepository struct {
	db *sql.DB
}

func NewPasswordCredentialRepository(db *sql.DB) PasswordCredentialRepository {
	return &postgresPasswordCredentialRepository{db: db}
}

func (r *postgresPasswordCredentialRepository) Create(ctx context.Context, userID, hash, salt string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_credential (user_id, password_hash, password_salt) VALUES ($1, $2, $3)`,
		userID, hash, salt,
	)
	return err
}

func (r *postgresPasswordCredentialRepository) GetByUserID(ctx context.Context, userID string) (*PasswordCredential, error) {
	cred := &PasswordCredential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, password_hash, password_salt FROM password_credential WHERE user_id = $1`,
		userID,
	).Scan(&cred.UserID, &cred.PasswordHash, &cred.PasswordSalt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cred, nil
}
