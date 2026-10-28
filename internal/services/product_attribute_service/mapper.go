package product_attribute_service

import (
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
	product_attribute2 "github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
)

func MapToResult(attr *product_attribute.Attribute) *product_attribute2.AttributeResult {
	if attr == nil {
		return nil
	}

	valResults := make([]product_attribute2.AttributeValueResult, 0, len(attr.Values()))
	for _, v := range attr.Values() {
		valResults = append(valResults, product_attribute2.AttributeValueResult{
			ID:          v.ID(),
			AttributeID: v.AttributeID(),
			Name:        v.Name(),
			SortOrder:   v.SortOrder(),
		})
	}

	return &product_attribute2.AttributeResult{
		ID:        attr.ID(),
		Name:      attr.Name(),
		SortOrder: attr.SortOrder(),
		Values:    valResults,
	}
}

func MapToResultList(attrs []*product_attribute.Attribute) []product_attribute2.AttributeResult {
	if attrs == nil {
		return nil
	}

	results := make([]product_attribute2.AttributeResult, len(attrs))
	for i, attr := range attrs {
		results[i] = *MapToResult(attr)
	}
	return results
}
