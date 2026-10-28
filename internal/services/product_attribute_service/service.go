package product_attribute_service

import (
	"context"
	"errors"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
)

type ProductAttributeService struct {
	repo port.ProductAttributeRepository
}

func NewProductAttributeService(repo port.ProductAttributeRepository) port.ProductAttributeService {
	return &ProductAttributeService{repo: repo}
}

func (s ProductAttributeService) Create(ctx context.Context, attr *domain.Attribute) (*domain.Attribute, error) {
	if attr == nil {
		return nil, errors.New("nil product attribute")
	}
	return s.repo.Create(ctx, attr)
}

func (s ProductAttributeService) Update(ctx context.Context, attr *domain.Attribute) (*domain.Attribute, error) {
	if attr == nil {
		return nil, errors.New("attribute is nil")
	}

	existing, err := s.repo.GetByID(ctx, attr.ID())
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("attribute not found")
	}

	return s.repo.Update(ctx, attr)
}

func (s ProductAttributeService) GetByID(ctx context.Context, id uint64) (*domain.Attribute, error) {
	if id == 0 {
		return nil, errors.New("invalid id")
	}

	return s.repo.GetByID(ctx, id)
}

func (s ProductAttributeService) List(ctx context.Context, q port.ListProductAttributeQuery) ([]*domain.Attribute, error) {
	if q.Limit <= 0 {
		q.Limit = 20
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	if q.Offset < 0 {
		q.Offset = 0
	}

	return s.repo.List(ctx, q)
}

func (s ProductAttributeService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("invalid id")
	}

	exists, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if exists == nil {
		return errors.New("not found")
	}

	return s.repo.Delete(ctx, id)
}
