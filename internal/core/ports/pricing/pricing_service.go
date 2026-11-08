package pricing

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
)

type CreatePricingItemDTO struct {
	ProductID          uint64  `json:"product_id"`
	Price              int64   `json:"price"`
	AllowPriceChange   bool    `json:"allow_price_change"`
	MaxDiscountPercent float64 `json:"max_discount_percent"`
}

type CreateAnnouncementDTO struct {
	Title        string                 `json:"title"`
	Code         string                 `json:"code"`
	RegisteredAt time.Time              `json:"registered_at"`
	Items        []CreatePricingItemDTO `json:"items"`
}

type PricingService interface {
	CreateAnnouncement(ctx context.Context, dto CreateAnnouncementDTO) (*pricing.Announcement, error)
	GetAnnouncementByID(ctx context.Context, id uint64) (*pricing.Announcement, error)
	ApproveAnnouncement(ctx context.Context, id uint64) error
	RejectAnnouncement(ctx context.Context, id uint64) error
	ListAnnouncements(ctx context.Context, filter ListAnnouncementsFilter) ([]*pricing.Announcement, int64, error)
	GetActiveProductPrice(ctx context.Context, productID uint64) (*ActiveProductPrice, error)
	GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]ActiveProductPrice, error)
}
