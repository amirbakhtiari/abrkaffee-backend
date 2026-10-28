package category

import "github.com/amirbakhtiari/abrcoffee/internal/core/ports"

type CategoryResponse struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Priority int    `json:"priority"`
	IsActive bool   `json:"is_active"`
}

func ToResponse(dto *ports.CategoryResponse) CategoryResponse {
	return CategoryResponse{
		ID:       dto.ID,
		Name:     dto.Name,
		Slug:     dto.Slug,
		Priority: dto.Priority,
		IsActive: dto.IsActive,
	}
}
