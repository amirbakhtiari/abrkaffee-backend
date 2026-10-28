package db_models

import "time"

type AccountDB struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	UserID       uint64 `gorm:"not null;uniqueIndex"`
	Username     string `gorm:"type:varchar(150);not null;uniqueIndex"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	User UserDB `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (AccountDB) TableName() string {
	return "accounts"
}
