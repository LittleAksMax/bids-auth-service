package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/LittleAksMax/bids-auth-service/internal/contracts"
	"github.com/LittleAksMax/bids-auth-service/internal/repository"
	"github.com/LittleAksMax/bids-util/passwords"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username or email already exists")
)

// AuthService handles authentication business logic.
type AuthService interface {
	Register(ctx context.Context, username, email, password, role string) (*contracts.UserDTO, error)
	Login(ctx context.Context, email, password string) (*contracts.UserDTO, error)
}

// authService implements AuthService.
type authService struct {
	pool         *sql.DB
	tokenService TokenService // NOTE: this breaks the
	userRepo     repository.UserRepository
	credRepo     repository.PasswordCredentialRepository
	pepper       string
}

// NewAuthService creates a new authentication service.
func NewAuthService(pool *sql.DB, userRepo repository.UserRepository, credRepo repository.PasswordCredentialRepository, pepper string) AuthService {
	return &authService{
		pool:     pool,
		userRepo: userRepo,
		credRepo: credRepo,
		pepper:   pepper,
	}
}

// Register creates a new user account and returns tokens for immediate use.
func (s *authService) Register(ctx context.Context, username, email, password, role string) (*contracts.UserDTO, error) {
	// Check uniqueness
	existingUsername, _ := s.userRepo.FindByUsername(ctx, s.pool, username)
	if existingUsername != nil {
		return nil, ErrUserExists
	}
	existingEmail, _ := s.userRepo.FindByEmail(ctx, s.pool, email)
	if existingEmail != nil {
		return nil, ErrUserExists
	}

	// Create transaction used for registration process to maintain consistency
	tx, err := s.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Ensure transaction is rolled back if any step fails (safe to do if commit has already happened)
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				// already committed or rolled back; no-op
				return
			}
			log.Printf("couldn't rollback transaction: %v\n", err)
		}
	}()

	// Create user in repository
	user, err := s.userRepo.Create(ctx, tx, username, email, role)
	if err != nil {
		return nil, ErrUserExists
	}

	// Store password credential
	salt, hash, err := passwords.HashPassword(password, s.pepper, passwords.DefaultParams)
	if err != nil {
		return nil, err
	}
	err = s.credRepo.Create(ctx, tx, user.ID, hash, salt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user.ToDTO(), nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*contracts.UserDTO, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, s.pool, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Get password credential
	creds, err := s.credRepo.GetByUserID(ctx, s.pool, user.ID)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	ok, err := passwords.VerifyPassword(password, s.pepper, creds.PasswordSalt, creds.PasswordHash, passwords.DefaultParams)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	return user.ToDTO(), nil
}
