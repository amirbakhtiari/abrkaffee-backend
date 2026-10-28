package db_models

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/auth"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

type UserDB struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(50);not null"`
	Family    string `gorm:"type:varchar(50);not null"`
	Email     string `gorm:"type:varchar(150);not null"`
	Mobile    string `gorm:"type:varchar(20);not null"`
	IsActive  bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserDB) TableName() string {
	return "users"
}

// FromAccount converts the profile of an account into a user database model.
func FromAccount(account *auth.Account) *UserDB {
	profile := account.Profile

	return &UserDB{
		ID:        profile.ID(),
		Name:      profile.Name(),
		Family:    profile.Family(),
		Email:     profile.Email(),
		Mobile:    profile.Mobile(),
		IsActive:  profile.IsActive(),
		CreatedAt: profile.CreatedAt(),
		UpdatedAt: profile.UpdatedAt(),
	}
}

// ToAccount converts a user database model into an account with its profile.
func (u *UserDB) ToAccount() *auth.Account {
	profile := user.Rehydrate(
		u.ID,
		u.Name,
		u.Family,
		u.Email,
		u.Mobile,
		nil,
		u.IsActive,
		u.CreatedAt,
		u.UpdatedAt,
	)

	return &auth.Account{
		UserID:  u.ID,
		Profile: profile,
	}
}
