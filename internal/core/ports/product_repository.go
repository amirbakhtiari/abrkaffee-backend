package ports

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
)

type ProductListQuery struct {
	Page       int
	PageSize   int
	Search     string
	IsActive   *bool
	CategoryID uint64
}

type ProductRepository interface {
	Create(ctx context.Context, product *product.Product) (*product.Product, error)
	GetByID(ctx context.Context, id uint64) (*product.Product, error)
	List(ctx context.Context, query ProductListQuery) ([]*product.Product, int64, error)
	Update(ctx context.Context, product *product.Product) (*product.Product, error)
	Delete(ctx context.Context, id uint64) error
	UpdateStatus(ctx context.Context, id uint64, isActive bool, updatedAt time.Time) error
}
