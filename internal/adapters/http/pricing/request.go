package pricing

import "time"

type CreatePricingItemRequest struct {
	ProductID          uint64  `json:"product_id" binding:"required"`
	Price              int64   `json:"price" binding:"required,gt=0"`
	AllowPriceChange   bool    `json:"allow_price_change"`
	MaxDiscountPercent float64 `json:"max_discount_percent" binding:"gte=0,lte=100"`
}

type CreateAnnouncementRequest struct {
	Title        string                     `json:"title" binding:"required"`
	Code         string                     `json:"code" binding:"required"`
	RegisteredAt *time.Time                 `json:"registered_at"`
	Items        []CreatePricingItemRequest `json:"items" binding:"required,min=1,dive"`
}
