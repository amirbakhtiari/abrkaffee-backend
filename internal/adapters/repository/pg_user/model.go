package pg_user

import "time"

type UserModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	Family    string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	Mobile    string    `gorm:"size:20;uniqueIndex;not null"`
	IsActive  bool      `gorm:"default:true;index"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time `gorm:"index"`
}
