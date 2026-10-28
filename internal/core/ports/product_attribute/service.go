package product_attribute

import (
	"context"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
)

type ProductAttributeService interface {
	Create(ctx context.Context, attr *domain.Attribute) (*domain.Attribute, error)
	Update(ctx context.Context, attr *domain.Attribute) (*domain.Attribute, error)
	GetByID(ctx context.Context, id uint64) (*domain.Attribute, error)
	List(ctx context.Context, q ListProductAttributeQuery) ([]*domain.Attribute, error)
	Delete(ctx context.Context, id uint64) error
}
