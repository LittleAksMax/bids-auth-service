package repository

import (
	"context"
	"database/sql"
)

// User represents a user entity.
type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
}

// UserRepository defines operations for user data access.
type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (userID string, err error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, userID string) (*User, error)
}

// postgresUserRepository implements UserRepository using PostgreSQL.
type postgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new Postgres-backed user repository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// Create inserts a new user and returns the generated ID.
func (r *postgresUserRepository) Create(ctx context.Context, username, email, passwordHash string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		username, email, passwordHash,
	).Scan(&userID)
	return userID, err
}

// FindByUsername retrieves a user by username.
func (r *postgresUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID retrieves a user by ID.
func (r *postgresUserRepository) FindByID(ctx context.Context, userID string) (*User, error) {
	user := &User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
