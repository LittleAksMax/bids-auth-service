package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/LittleAksMax/bids-auth-service/internal/contracts"
	"github.com/google/uuid"
)

// UserRepository defines operations for user data access.
type UserRepository interface {
	Create(ctx context.Context, tx *sql.Tx, username, email, role string) (*contracts.User, error)
	FindByUsername(ctx context.Context, db *sql.DB, username string) (*contracts.User, error)
	FindByID(ctx context.Context, db *sql.DB, userID uuid.UUID) (*contracts.User, error)
	FindByEmail(ctx context.Context, db *sql.DB, email string) (*contracts.User, error)
}

// postgresUserRepository implements UserRepository using PostgreSQL.
type postgresUserRepository struct {
}

// NewUserRepository creates a new Postgres-backed user repository.
func NewUserRepository() UserRepository {
	return &postgresUserRepository{}
}

// Create inserts a new user and returns the generated ID.
func (r *postgresUserRepository) Create(ctx context.Context, tx *sql.Tx, username, email, role string) (*contracts.User, error) {
	user := &contracts.User{}
	err := tx.QueryRowContext(ctx,
		`INSERT INTO users (username, email, role) VALUES ($1, $2, $3) RETURNING id, username, email, created_at, updated_at, "role"`,
		username, email, role,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Role)
	return user, err
}

// FindByUsername retrieves a user by username.
func (r *postgresUserRepository) FindByUsername(ctx context.Context, db *sql.DB, username string) (*contracts.User, error) {
	user := &contracts.User{}
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, created_at, updated_at, role FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Role)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID retrieves a user by ID.
func (r *postgresUserRepository) FindByID(ctx context.Context, db *sql.DB, userID uuid.UUID) (*contracts.User, error) {
	user := &contracts.User{}
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, created_at, updated_at, role FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Role)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByEmail retrieves a user by email
func (r *postgresUserRepository) FindByEmail(ctx context.Context, db *sql.DB, email string) (*contracts.User, error) {
	user := &contracts.User{}
	err := db.QueryRowContext(ctx,
		`SELECT id, username, email, created_at, updated_at, role FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
