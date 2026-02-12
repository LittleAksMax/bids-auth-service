package contracts

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity.
type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Role      string
}

func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Role:      u.Role,
	}
}

// RefreshToken represents a refresh token stored in the database.
type RefreshToken struct {
	TokenID           uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	IssuedAt          time.Time
	ExpiresAt         time.Time
	RevokedAt         *time.Time
	ReplacedByTokenID *uuid.UUID
}

type PasswordCredential struct {
	UserID       uuid.UUID
	PasswordHash string
	PasswordSalt string
}
