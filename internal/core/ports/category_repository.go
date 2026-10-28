package ports

import (
	"context"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *category.Category) (*category.Category, error)
	GetByID(ctx context.Context, id uint64) (*category.Category, error)
	GetBySlug(ctx context.Context, slug string) (*category.Category, error)
	GetAll(ctx context.Context, skip uint64, take uint64) ([]*category.Category, error)
	Update(ctx context.Context, category *category.Category) error
	Delete(ctx context.Context, id uint64) error
}
