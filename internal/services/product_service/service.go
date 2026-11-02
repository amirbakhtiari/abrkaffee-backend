package product_service

import (
	"context"
	"errors"
	"strings"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

var ErrProductRequestRequired = errors.New("product request is required")

type productService struct {
	repo         ports.ProductRepository
	categoryRepo ports.CategoryRepository
	clock        product.Clock
}

func NewProductService(
	repo ports.ProductRepository,
	categoryRepo ports.CategoryRepository,
	clock product.Clock,
) ports.ProductService {
	return &productService{repo: repo, categoryRepo: categoryRepo, clock: clock}
}

func (s *productService) CreateProduct(ctx context.Context, req *ports.CreateProductDTO) (*ports.ProductResponse, error) {
	if req == nil {
		return nil, ErrProductRequestRequired
	}
	if err := s.validateCategories(ctx, req.CategoryIDs); err != nil {
		return nil, err
	}

	item, err := product.NewProduct(
		req.Name, req.Code, req.SKU, req.Barcode, req.Description,
		req.IsActive, req.CategoryIDs, s.clock,
	)
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	return toServiceResponse(saved), nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uint64, req *ports.UpdateProductDTO) (*ports.ProductResponse, error) {
	if req == nil {
		return nil, ErrProductRequestRequired
	}
	if err := s.validateCategories(ctx, req.CategoryIDs); err != nil {
		return nil, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := item.Edit(
		req.Name, req.Code, req.SKU, req.Barcode, req.Description,
		req.CategoryIDs, s.clock,
	); err != nil {
		return nil, err
	}

	if req.IsActive != nil {
		if *req.IsActive {
			item.Activate(s.clock)
		} else {
			item.Deactivate(s.clock)
		}
	}

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, err
	}
	return toServiceResponse(updated), nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uint64) error {
	if id == 0 {
		return product.ErrProductNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *productService) GetProductByID(ctx context.Context, id uint64) (*ports.ProductResponse, error) {
	if id == 0 {
		return nil, product.ErrProductNotFound
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toServiceResponse(item), nil
}

func (s *productService) ListProducts(ctx context.Context, dto ports.ListProductsDTO) ([]*ports.ProductResponse, int64, error) {
	if dto.Page < 1 {
		dto.Page = 1
	}
	if dto.PageSize < 1 {
		dto.PageSize = 10
	}
	if dto.PageSize > 100 {
		dto.PageSize = 100
	}
	dto.Search = strings.TrimSpace(dto.Search)

	items, total, err := s.repo.List(ctx, ports.ProductListQuery{
		Page: dto.Page, PageSize: dto.PageSize, Search: dto.Search,
		IsActive: dto.IsActive, CategoryID: dto.CategoryID,
	})
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*ports.ProductResponse, len(items))
	for i, item := range items {
		responses[i] = toServiceResponse(item)
	}
	return responses, total, nil
}

func (s *productService) validateCategories(ctx context.Context, ids []uint64) error {
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return product.ErrProductCategoryInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}

		cat, err := s.categoryRepo.GetByID(ctx, id)
		if errors.Is(err, category.ErrCategoryNotFound) || (err == nil && cat == nil) {
			return product.ErrProductCategoryInvalid
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func toServiceResponse(item *product.Product) *ports.ProductResponse {
	if item == nil {
		return nil
	}
	return &ports.ProductResponse{
		ID: item.ID(), Name: item.Name(), Code: item.Code(), SKU: item.SKU(),
		Barcode: item.Barcode(), Description: item.Description(), IsActive: item.IsActive(),
		CategoryIDs: item.CategoryIDs(), CreatedAt: item.CreatedAt(), UpdatedAt: item.UpdatedAt(),
	}
}
