package pricing

import (
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
)

type PricingItemResponse struct {
	ID                 uint64  `json:"id"`
	ProductID          uint64  `json:"product_id"`
	ProductName        string  `json:"product_name"`
	ProductCode        string  `json:"product_code"`
	Price              int64   `json:"price"`
	AllowPriceChange   bool    `json:"allow_price_change"`
	MaxDiscountPercent float64 `json:"max_discount_percent"`
}

type AnnouncementResponse struct {
	ID           uint64                `json:"id"`
	Title        string                `json:"title"`
	Code         string                `json:"code"`
	Status       string                `json:"status"`
	RegisteredAt time.Time             `json:"registered_at"`
	ApprovedAt   *time.Time            `json:"approved_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Items        []PricingItemResponse `json:"items"`
}

type AnnouncementSummaryResponse struct {
	ID           uint64     `json:"id"`
	Title        string     `json:"title"`
	Code         string     `json:"code"`
	Status       string     `json:"status"`
	RegisteredAt time.Time  `json:"registered_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ListAnnouncementsResponse struct {
	Items []AnnouncementSummaryResponse `json:"items"`
	Total int64                         `json:"total"`
	Page  int                           `json:"page"`
	Limit int                           `json:"limit"`
}

func ToResponse(a *domain.Announcement) AnnouncementResponse {
	items := make([]PricingItemResponse, 0, len(a.Items()))
	for _, it := range a.Items() {
		items = append(items, PricingItemResponse{
			ID:                 it.ID(),
			ProductID:          it.ProductID(),
			ProductName:        it.ProductName(),
			ProductCode:        it.ProductCode(),
			Price:              it.Price(),
			AllowPriceChange:   it.AllowPriceChange(),
			MaxDiscountPercent: it.MaxDiscountPercent(),
		})
	}

	return AnnouncementResponse{
		ID:           a.ID(),
		Title:        a.Title(),
		Code:         a.Code(),
		Status:       string(a.Status()),
		RegisteredAt: a.RegisteredAt(),
		ApprovedAt:   a.ApprovedAt(),
		CreatedAt:    a.CreatedAt(),
		UpdatedAt:    a.UpdatedAt(),
		Items:        items,
	}
}

func ToSummaryResponse(a *domain.Announcement) AnnouncementSummaryResponse {
	return AnnouncementSummaryResponse{
		ID:           a.ID(),
		Title:        a.Title(),
		Code:         a.Code(),
		Status:       string(a.Status()),
		RegisteredAt: a.RegisteredAt(),
		ApprovedAt:   a.ApprovedAt(),
		CreatedAt:    a.CreatedAt(),
		UpdatedAt:    a.UpdatedAt(),
	}
}
