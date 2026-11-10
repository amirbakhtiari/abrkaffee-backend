package ports

import (
	"context"
	"time"
)

type ProductPriceDTO struct {
	AnnouncementItemID uint64
	AnnouncementID     uint64
	Price              int64
	AllowPriceChange   bool
	MaxDiscountPercent float64
	ApprovedAt         time.Time
}

type ProductWithPriceDTO struct {
	Product ProductResponse
	Price   *ProductPriceDTO
}

type ListProductsWithPricesDTO struct {
	Page       int
	PageSize   int
	Search     string
	IsActive   *bool
	CategoryID uint64
}

type ProductPricingService interface {
	ListProductsWithPrices(
		ctx context.Context,
		dto ListProductsWithPricesDTO,
	) ([]*ProductWithPriceDTO, int64, error)
}
