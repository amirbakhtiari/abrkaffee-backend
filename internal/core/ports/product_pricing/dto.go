package product_pricing

import "time"

type ProductWithPrice struct {
	ProductID   uint64
	Name        string
	Code        string
	SKU         string
	Barcode     string
	Description string
	IsActive    bool
	CategoryIDs []uint64
	ImageURL    string

	Price              *int64
	AnnouncementItemID *uint64
	AnnouncementID     *uint64
	AllowPriceChange   bool
	MaxDiscountPercent float64
	PriceApprovedAt    *time.Time
}

type ListProductsWithPricesDTO struct {
	Page       int
	PageSize   int
	Search     string
	IsActive   *bool
	CategoryID uint64
}
