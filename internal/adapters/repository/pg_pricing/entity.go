package pg_pricing

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_product"
	"gorm.io/gorm"
)

type PriceAnnouncementEntity struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Title        string    `gorm:"type:varchar(255);not null"`
	Code         string    `gorm:"type:varchar(64);not null;uniqueIndex"`
	RegisteredAt time.Time `gorm:"not null"`
	Status       string    `gorm:"type:varchar(20);not null;default:PENDING;index"`
	ApprovedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	Items []PriceAnnouncementItemEntity `gorm:"foreignKey:AnnouncementID;constraint:OnDelete:CASCADE"`
}

func (PriceAnnouncementEntity) TableName() string {
	return "price_announcements"
}

type PriceAnnouncementItemEntity struct {
	ID                 uint64                   `gorm:"primaryKey;autoIncrement"`
	AnnouncementID     uint64                   `gorm:"not null;uniqueIndex:idx_announcement_product,priority:1"`
	ProductID          uint64                   `gorm:"not null;uniqueIndex:idx_announcement_product,priority:2;index"`
	Product            pg_product.ProductEntity `gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:RESTRICT"`
	Price              int64                    `gorm:"not null"`
	AllowPriceChange   bool                     `gorm:"not null;default:false"`
	MaxDiscountPercent float64                  `gorm:"type:numeric(5,2);not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PriceAnnouncementItemEntity) TableName() string {
	return "price_announcement_items"
}
