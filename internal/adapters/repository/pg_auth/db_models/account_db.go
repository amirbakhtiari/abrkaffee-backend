package db_models

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
)

type AccountDB struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	UserID       uint64 `gorm:"not null;uniqueIndex"`
	Username     string `gorm:"type:varchar(150);not null;uniqueIndex"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"`
	Role         string `gorm:"type:varchar(20);not null;default:Staff"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	User pg_user.UserModel `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (AccountDB) TableName() string {
	return "accounts"
}
