package product_attribute

import domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"

type AttributeResponse struct {
	ID        uint64              `json:"id"`
	Name      string              `json:"name"`
	SortOrder int                 `json:"sort_order"`
	Values    []AttributeValueRes `json:"values,omitempty"`
}

type AttributeValueRes struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type ListProductAttributeResponse struct {
	Items []AttributeResponse `json:"items"`
}

func toResponse(a *domain.Attribute) AttributeResponse {
	values := make([]AttributeValueRes, 0, len(a.Values()))

	for _, v := range a.Values() {
		values = append(values, AttributeValueRes{
			ID:        v.ID(),
			Name:      v.Name(),
			SortOrder: v.SortOrder(),
		})
	}

	return AttributeResponse{
		ID:        a.ID(),
		Name:      a.Name(),
		SortOrder: a.SortOrder(),
		Values:    values,
	}
}
