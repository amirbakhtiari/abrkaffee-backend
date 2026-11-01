package pg_category

import (
	"context"
	"errors"
	"fmt"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/amirbakhtiari/abrcoffee/pkg/dberrors"
	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) ports.CategoryRepository {
	return &pgRepository{
		db: db,
	}
}

func (r pgRepository) Create(
	ctx context.Context,
	cat *category.Category,
) (*category.Category, error) {

	if cat == nil {
		return nil, category.ErrCategoryRequired
	}

	entity := FromDomain(cat)

	if err := r.db.
		WithContext(ctx).
		Create(entity).Error; err != nil {
		return nil, dberrors.Translate(err)
	}

	return entity.ToDomain(), nil
}

func (r pgRepository) GetByID(
	ctx context.Context,
	id uint64,
) (*category.Category, error) {
	var entity CategoryEntity

	err := r.db.
		WithContext(ctx).
		First(&entity, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrCategoryNotFound
		}

		return nil, dberrors.Translate(err)
	}

	return entity.ToDomain(), nil
}

func (r pgRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*category.Category, error) {
	var entity CategoryEntity

	err := r.db.
		WithContext(ctx).
		First(&entity, "slug = ?", slug).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrCategoryNotFound
		}

		return nil, dberrors.Translate(err)
	}

	return entity.ToDomain(), nil
}

func (r pgRepository) GetAll(
	ctx context.Context,
	skip uint64,
	take uint64,
) ([]*category.Category, error) {
	maxInt := uint64(^uint(0) >> 1)

	if skip > maxInt || take > maxInt {
		return nil, fmt.Errorf("category pagination exceeds supported integer range")
	}

	var entities []*CategoryEntity

	err := r.db.
		WithContext(ctx).
		Order("priority ASC, id ASC").
		Offset(int(skip)).
		Limit(int(take)).
		Find(&entities).Error

	if err != nil {
		return nil, dberrors.Translate(err)
	}

	categories := make([]*category.Category, len(entities))

	for i, entity := range entities {
		categories[i] = entity.ToDomain()
	}

	return categories, nil
}

func (r pgRepository) Update(
	ctx context.Context,
	cat *category.Category,
) error {

	if cat == nil {
		return category.ErrCategoryRequired
	}

	entity := FromDomain(cat)

	result := r.db.
		WithContext(ctx).
		Model(&CategoryEntity{}).
		Where("id = ?", entity.ID).
		Updates(map[string]interface{}{
			"name":       entity.Name,
			"slug":       entity.Slug,
			"is_active":  entity.IsActive,
			"priority":   entity.Priority,
			"updated_at": entity.UpdatedAt,
		})

	if result.Error != nil {
		return dberrors.Translate(result.Error)
	}

	if result.RowsAffected == 0 {
		return category.ErrCategoryNotFound
	}

	return nil
}

func (r pgRepository) Delete(
	ctx context.Context,
	id uint64,
) error {
	result := r.db.
		WithContext(ctx).
		Delete(&CategoryEntity{}, id)

	if result.Error != nil {
		return dberrors.Translate(result.Error)
	}

	if result.RowsAffected == 0 {
		return category.ErrCategoryNotFound
	}

	return nil
}
