package pricing

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
)

type ActiveProductPrice struct {
	Price              int64
	AnnouncementItemID uint64
	AnnouncementID     uint64
	AllowPriceChange   bool
	MaxDiscountPercent float64
	ApprovedAt         time.Time
}

type ListAnnouncementsFilter struct {
	Status string
	Page   int
	Limit  int
}

type PricingRepository interface {
	Create(ctx context.Context, announcement *pricing.Announcement) (*pricing.Announcement, error)
	GetByID(ctx context.Context, id uint64) (*pricing.Announcement, error)
	UpdateStatus(ctx context.Context, announcement *pricing.Announcement) error
	List(ctx context.Context, filter ListAnnouncementsFilter) ([]*pricing.Announcement, int64, error)
	GetActivePrice(ctx context.Context, productID uint64) (*ActiveProductPrice, error)
	GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]ActiveProductPrice, error)
}
