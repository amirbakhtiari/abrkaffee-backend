package ports

import "context"

type CreateCategoryDTO struct {
	Name     string
	Slug     string
	Priority int
	IsActive bool
}

type UpdateCategoryDTO struct {
	Name     string
	Slug     string
	Priority int
	IsActive *bool
}

type CategoryResponse struct {
	ID       uint64
	Name     string
	Slug     string
	IsActive bool
	Priority int
}

type ListCategoriesDTO struct {
	Skip uint64
	Take uint64
}

type CategoryService interface {
	CreateCategory(ctx context.Context, req *CreateCategoryDTO) (*CategoryResponse, error)
	UpdateCategory(ctx context.Context, id uint64, req *UpdateCategoryDTO) (*CategoryResponse, error)
	DeleteCategory(ctx context.Context, id uint64) error
	GetCategoryByID(ctx context.Context, id uint64) (*CategoryResponse, error)
	ListCategories(ctx context.Context, dto ListCategoriesDTO) ([]*CategoryResponse, error)
}
