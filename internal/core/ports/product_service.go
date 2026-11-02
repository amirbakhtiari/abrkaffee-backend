package ports

import (
	"context"
	"time"
)

type CreateProductDTO struct {
	Name        string
	Code        string
	SKU         string
	Barcode     string
	Description string
	IsActive    bool
	CategoryIDs []uint64
}

type UpdateProductDTO struct {
	Name        string
	Code        string
	SKU         string
	Barcode     string
	Description string
	IsActive    *bool
	CategoryIDs []uint64
}

type ProductResponse struct {
	ID          uint64
	Name        string
	Code        string
	SKU         string
	Barcode     string
	Description string
	IsActive    bool
	CategoryIDs []uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProductsDTO struct {
	Page     int
	PageSize int
	Search   string
	IsActive *bool
}

type ProductService interface {
	CreateProduct(ctx context.Context, req *CreateProductDTO) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, id uint64, req *UpdateProductDTO) (*ProductResponse, error)
	DeleteProduct(ctx context.Context, id uint64) error
	GetProductByID(ctx context.Context, id uint64) (*ProductResponse, error)
	ListProducts(ctx context.Context, dto ListProductsDTO) ([]*ProductResponse, int64, error)
}
