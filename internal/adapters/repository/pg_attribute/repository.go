package pg_attribute

import (
	"context"

	attribute "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
	"gorm.io/gorm"
)

type AttributeRepository struct {
	db *gorm.DB
}

func NewAttributeRepository(db *gorm.DB) port.ProductAttributeRepository {
	return &AttributeRepository{db: db}
}

func (r *AttributeRepository) Create(ctx context.Context, attr *attribute.Attribute) (*attribute.Attribute, error) {
	entity := FromDomain(attr)

	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return nil, err
	}
	return ToDomain(entity), nil
}

func (r *AttributeRepository) Update(ctx context.Context, attr *attribute.Attribute) (*attribute.Attribute, error) {
	dbModel := FromDomain(attr)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(dbModel).Error; err != nil {
			return err
		}

		if err := tx.
			Where("attribute_id = ?", attr.ID()).
			Delete(&ProductAttributeValueModel{}).Error; err != nil {
			return err
		}

		if len(dbModel.Values) > 0 {
			if err := tx.Create(&dbModel.Values).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, attr.ID())
}

func (r *AttributeRepository) GetByID(ctx context.Context, id uint64) (*attribute.Attribute, error) {
	var entity ProductAttributeModel
	err := r.db.WithContext(ctx).Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return ToDomain(&entity), nil
}

func (r *AttributeRepository) List(ctx context.Context, q port.ListProductAttributeQuery) ([]*attribute.Attribute, int64, error) {
	var (
		entities []ProductAttributeModel
		total    int64
	)

	db := r.db.WithContext(ctx).Model(&ProductAttributeModel{})

	if q.Search != "" {
		db = db.Where("name ILIKE ?", "%"+q.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Order("sort_order ASC").
		Limit(q.Limit).
		Offset(q.Offset).
		Find(&entities).
		Error

	if err != nil {
		return nil, 0, err
	}

	result := make([]*attribute.Attribute, len(entities))
	for i := range entities {
		result[i] = ToDomain(&entities[i])
	}

	return result, total, nil
}

func (r *AttributeRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&ProductAttributeModel{}, id).Error
}
