package pg_attribute

import (
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
)

func FromDomain(attr *product_attribute.Attribute) *ProductAttributeModel {
	if attr == nil {
		return nil
	}

	dbValues := make([]ProductAttributeValueModel, 0, len(attr.Values()))
	for _, v := range attr.Values() {
		dbValues = append(dbValues, ProductAttributeValueModel{
			ID:          v.ID(),
			AttributeID: v.AttributeID(),
			Name:        v.Name(),
			SortOrder:   v.SortOrder(),
			CreatedAt:   v.CreatedAt(),
		})
	}

	return &ProductAttributeModel{
		ID:        attr.ID(),
		Name:      attr.Name(),
		SortOrder: attr.SortOrder(),
		CreatedAt: attr.CreatedAt(),
		Values:    dbValues,
	}
}

func ToDomain(db *ProductAttributeModel) *product_attribute.Attribute {
	if db == nil {
		return nil
	}

	attr, _ := product_attribute.NewAttribute(db.Name, db.SortOrder)

	for _, valEntity := range db.Values {
		val, _ := product_attribute.NewValue(valEntity.AttributeID, valEntity.Name, valEntity.SortOrder)
		err := attr.AddValue(val)
		if err != nil {
			return nil
		}
	}

	return attr
}
