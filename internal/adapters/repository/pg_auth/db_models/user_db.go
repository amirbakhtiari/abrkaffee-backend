package db_models

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
)

type UserDB struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement"`
	Name      string     `gorm:"type:varchar(50);not null"`
	Family    string     `gorm:"type:varchar(50);not null"`
	Email     string     `gorm:"type:varchar(50);not null"`
	Username  string     `gorm:"type:varchar(150);not null;index,unique"`
	Password  string     `gorm:"type:varchar(255);not null;"`
	Mobile    string     `gorm:"type:varchar(20);not null;index,unique"`
	Birthdate *time.Time `gorm:"type:date"`
	IsActive  bool       `gorm:"type:boolean;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

func (UserDB) TableName() string {
	return "users"
}

func FromDomain(u *user.User) *UserDB {
	return &UserDB{
		ID:        u.ID,
		Name:      u.Name,
		Family:    u.Family,
		Email:     u.Email,
		Username:  u.Username,
		Password:  u.Password,
		Mobile:    u.Mobile,
		Birthdate: u.Birthdate,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
func (u *UserDB) ToDomain() *user.User {
	return &user.User{
		ID:        u.ID,
		Name:      u.Name,
		Family:    u.Family,
		Username:  u.Username,
		Email:     u.Email,
		Password:  u.Password,
		Mobile:    u.Mobile,
		Birthdate: u.Birthdate,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
