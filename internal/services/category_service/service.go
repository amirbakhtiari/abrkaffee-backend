package category_service

import (
	"context"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)

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
	cat, err := category.NewCategory(req.Name, req.Slug, req.IsActive, req.Priority, s.clock)
	if err != nil {
		return nil, err
	}
	saveCat, err := s.repo.Create(ctx, cat)
	if err != nil {
		return nil, err
	}
	return toServiceResponse(saveCat), nil
}

func (s categoryService) UpdateCategory(ctx context.Context, id uint64, req *ports.UpdateCategoryDTO) (*ports.CategoryResponse, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	err = cat.Edit(req.Name, req.Slug, req.Priority, s.clock)
	if err != nil {
		return nil, err
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
	if dto.Skip == 0 || dto.Take == 0 {
		dto.Take = 10
	}
	cats, err := s.repo.GetAll(ctx, dto.Skip, dto.Take)
	if err != nil {
		return nil, err
	}
	res := make([]*ports.CategoryResponse, len(cats))
	for i, cat := range cats {
		res[i] = toServiceResponse(cat)
	}
	return res, nil
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
