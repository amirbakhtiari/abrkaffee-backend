package product_pricing

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

type ProductPriceResponse struct {
	AnnouncementItemID uint64    `json:"announcement_item_id"`
	AnnouncementID     uint64    `json:"announcement_id"`
	Price              int64     `json:"price"`
	AllowPriceChange   bool      `json:"allow_price_change"`
	MaxDiscountPercent float64   `json:"max_discount_percent"`
	ApprovedAt         time.Time `json:"approved_at"`
}

type ProductWithPriceResponse struct {
	ID          uint64                `json:"id"`
	Name        string                `json:"name"`
	Code        string                `json:"code"`
	SKU         string                `json:"sku,omitempty"`
	Barcode     string                `json:"barcode,omitempty"`
	Description string                `json:"description"`
	IsActive    bool                  `json:"is_active"`
	CategoryIDs []uint64              `json:"category_ids"`
	ImageURL    string                `json:"image_url"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Price       *ProductPriceResponse `json:"price"`
}

type PublicProductPriceResponse struct {
	Price int64 `json:"price"`
}

type PublicProductWithPriceResponse struct {
	ID          uint64                      `json:"id"`
	Name        string                      `json:"name"`
	Code        string                      `json:"code"`
	Description string                      `json:"description"`
	CategoryIDs []uint64                    `json:"category_ids"`
	ImageURL    string                      `json:"image_url"`
	Price       *PublicProductPriceResponse `json:"price"`
}

type ListProductsWithPricesResponse struct {
	Items    []ProductWithPriceResponse `json:"items"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

type PublicListProductsWithPricesResponse struct {
	Items    []PublicProductWithPriceResponse `json:"items"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"page_size"`
}

func toPriceResponse(dto *ports.ProductPriceDTO) *ProductPriceResponse {
	if dto == nil {
		return nil
	}

	return &ProductPriceResponse{
		AnnouncementItemID: dto.AnnouncementItemID,
		AnnouncementID:     dto.AnnouncementID,
		Price:              dto.Price,
		AllowPriceChange:   dto.AllowPriceChange,
		MaxDiscountPercent: dto.MaxDiscountPercent,
		ApprovedAt:         dto.ApprovedAt,
	}
}

func toProductWithPriceResponse(
	dto *ports.ProductWithPriceDTO,
) ProductWithPriceResponse {
	return ProductWithPriceResponse{
		ID:          dto.Product.ID,
		Name:        dto.Product.Name,
		Code:        dto.Product.Code,
		SKU:         dto.Product.SKU,
		Barcode:     dto.Product.Barcode,
		Description: dto.Product.Description,
		IsActive:    dto.Product.IsActive,
		CategoryIDs: dto.Product.CategoryIDs,
		ImageURL:    dto.Product.ImageURL,
		CreatedAt:   dto.Product.CreatedAt,
		UpdatedAt:   dto.Product.UpdatedAt,
		Price:       toPriceResponse(dto.Price),
	}
}

func toPublicProductWithPriceResponse(
	dto *ports.ProductWithPriceDTO,
) PublicProductWithPriceResponse {
	return PublicProductWithPriceResponse{
		ID:          dto.Product.ID,
		Name:        dto.Product.Name,
		Code:        dto.Product.Code,
		Description: dto.Product.Description,
		CategoryIDs: dto.Product.CategoryIDs,
		ImageURL:    dto.Product.ImageURL,
		Price:       toPublicPriceResponse(dto.Price),
	}
}

func toPublicPriceResponse(dto *ports.ProductPriceDTO) *PublicProductPriceResponse {
	if dto == nil {
		return nil
	}

	return &PublicProductPriceResponse{
		Price: dto.Price,
	}
}
