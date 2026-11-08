package product

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

type ProductResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	SKU         string    `json:"sku,omitempty"`
	Barcode     string    `json:"barcode,omitempty"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CategoryIDs []uint64  `json:"category_ids"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicProductResponse struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Description string   `json:"description"`
	CategoryIDs []uint64 `json:"category_ids"`
	ImageURL    string   `json:"image_url"`
}

type ListResponse struct {
	Items    []ProductResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type PublicListResponse struct {
	Items    []PublicProductResponse `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

func ToResponse(dto *ports.ProductResponse) ProductResponse {
	return ProductResponse{
		ID:          dto.ID,
		Name:        dto.Name,
		Code:        dto.Code,
		SKU:         dto.SKU,
		Barcode:     dto.Barcode,
		Description: dto.Description,
		IsActive:    dto.IsActive,
		CategoryIDs: dto.CategoryIDs,
		ImageURL:    dto.ImageURL,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func ToPublicResponse(dto *ports.ProductResponse) PublicProductResponse {
	return PublicProductResponse{
		ID:          dto.ID,
		Name:        dto.Name,
		Code:        dto.Code,
		Description: dto.Description,
		CategoryIDs: dto.CategoryIDs,
		ImageURL:    dto.ImageURL,
	}
}
