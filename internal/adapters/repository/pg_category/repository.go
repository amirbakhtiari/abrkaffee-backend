package pg_category

import (
	"context"
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/amirbakhtiari/abrcoffee/pkg/dberrors"
	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) ports.CategoryRepository {
	return &pgRepository{db: db}
}

func (r pgRepository) Create(ctx context.Context, category *category.Category) (*category.Category, error) {
	entity := FromDomain(category)
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, dberrors.Translate(err)
	}
	return entity.ToDomain(), nil
}

func (r pgRepository) GetByID(ctx context.Context, id uint64) (*category.Category, error) {
	var entity CategoryEntity
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrCategoryNotFound
		}
		return nil, err
	}
	return entity.ToDomain(), nil
}

func (r pgRepository) GetBySlug(ctx context.Context, slug string) (*category.Category, error) {
	var entity CategoryEntity
	if err := r.db.WithContext(ctx).First(&entity, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrCategoryNotFound
		}
		return nil, err
	}
	return entity.ToDomain(), nil
}

func (r pgRepository) GetAll(ctx context.Context, skip uint64, take uint64) ([]*category.Category, error) {
	var entities []*CategoryEntity
	if err := r.db.WithContext(ctx).
		Order("priority asc, id asc").
		Offset(int(skip)).
		Limit(int(take)).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	categories := make([]*category.Category, len(entities))
	for i, entity := range entities {
		categories[i] = entity.ToDomain()
	}
	return categories, nil
}

func (r pgRepository) Update(ctx context.Context, category *category.Category) error {
	entity := FromDomain(category)
	return r.db.WithContext(ctx).Model(entity).Select("*").Updates(entity).Error
}

func (r pgRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&CategoryEntity{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return category.ErrCategoryNotFound
	}
	return nil
}
