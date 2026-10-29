package auth

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

type Account struct {
	ID           uint64
	UserID       uint64
	Profile      *user.User
	Username     string
	PasswordHash string
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAccount(
	userID uint64,
	profile *user.User,
	username string,
	passwordHash string,
	role string,
	now time.Time,
) *Account {
	return &Account{
		UserID:       userID,
		Profile:      profile,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
