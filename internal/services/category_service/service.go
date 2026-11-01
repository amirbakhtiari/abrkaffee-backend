package category_service

import (
	"context"
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

var ErrCategoryRequestRequired = errors.New("category request is required")

type categoryService struct {
	repo  ports.CategoryRepository
	clock category.Clock
}

func NewCategoryService(repo ports.CategoryRepository, clock category.Clock) ports.CategoryService {
	return &categoryService{
		repo:  repo,
		clock: clock,
	}
}

func (s categoryService) CreateCategory(ctx context.Context, req *ports.CreateCategoryDTO) (*ports.CategoryResponse, error) {
	if req == nil {
		return nil, ErrCategoryRequestRequired
	}

	cat, err := category.NewCategory(req.Name, req.Slug, req.IsActive, req.Priority, s.clock)
	if err != nil {
		return nil, err
	}

	savedCategory, err := s.repo.Create(ctx, cat)
	if err != nil {
		return nil, err
	}

	return toServiceResponse(savedCategory), nil
}

func (s categoryService) UpdateCategory(ctx context.Context, id uint64, req *ports.UpdateCategoryDTO) (*ports.CategoryResponse, error) {
	if req == nil {
		return nil, ErrCategoryRequestRequired
	}

	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := cat.Edit(req.Name, req.Slug, req.Priority, s.clock); err != nil {
		return nil, err
	}

	if req.IsActive != nil {
		if *req.IsActive {
			cat.Activate(s.clock)
		} else {
			cat.Deactivate(s.clock)
		}
	}

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	return toServiceResponse(cat), nil
}

func (s categoryService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s categoryService) GetCategoryByID(ctx context.Context, id uint64) (*ports.CategoryResponse, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toServiceResponse(cat), nil
}

func (s categoryService) ListCategories(ctx context.Context, dto ports.ListCategoriesDTO) ([]*ports.CategoryResponse, error) {
	if dto.Take == 0 {
		dto.Take = 10
	}

	categories, err := s.repo.GetAll(ctx, dto.Skip, dto.Take)
	if err != nil {
		return nil, err
	}

	responses := make([]*ports.CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = toServiceResponse(cat)
	}

	return responses, nil
}

func toServiceResponse(c *category.Category) *ports.CategoryResponse {
	return &ports.CategoryResponse{
		ID:       c.ID(),
		Name:     c.Name(),
		Slug:     c.Slug(),
		IsActive: c.IsActive(),
		Priority: c.Priority(),
	}
}
