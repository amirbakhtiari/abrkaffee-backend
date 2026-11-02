package pg_product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type pgRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) ports.ProductRepository {
	return &pgRepository{db: db}
}

func (r *pgRepository) Create(ctx context.Context, item *product.Product) (*product.Product, error) {
	if item == nil {
		return nil, product.ErrProductRequired
	}

	entity := FromDomain(item)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return translateError(err)
		}
		return replaceCategories(tx, entity.ID, item.CategoryIDs())
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, entity.ID)
}

func (r *pgRepository) GetByID(ctx context.Context, id uint64) (*product.Product, error) {
	var entity ProductEntity
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, product.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	ids, err := loadCategoryIDs(r.db.WithContext(ctx), []uint64{entity.ID})
	if err != nil {
		return nil, err
	}
	return entity.ToDomain(ids[entity.ID]), nil
}

func (r *pgRepository) List(ctx context.Context, query ports.ProductListQuery) ([]*product.Product, int64, error) {
	var (
		entities []ProductEntity
		total    int64
	)

	db := r.db.WithContext(ctx).Model(&ProductEntity{})
	if query.Search != "" {
		pattern := "%" + strings.TrimSpace(query.Search) + "%"
		db = db.Where(
			"name ILIKE ? OR code ILIKE ? OR sku ILIKE ? OR barcode ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC, id DESC").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	ids := make([]uint64, len(entities))
	for i := range entities {
		ids[i] = entities[i].ID
	}
	categoryIDs, err := loadCategoryIDs(r.db.WithContext(ctx), ids)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*product.Product, len(entities))
	for i := range entities {
		result[i] = entities[i].ToDomain(categoryIDs[entities[i].ID])
	}
	return result, total, nil
}

func (r *pgRepository) Update(ctx context.Context, item *product.Product) (*product.Product, error) {
	if item == nil {
		return nil, product.ErrProductRequired
	}

	entity := FromDomain(item)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&ProductEntity{}).Where("id = ?", item.ID()).Updates(map[string]interface{}{
			"name": entity.Name, "code": entity.Code, "sku": entity.SKU,
			"barcode": entity.Barcode, "description": entity.Description,
			"is_active": entity.IsActive, "updated_at": entity.UpdatedAt,
		})
		if result.Error != nil {
			return translateError(result.Error)
		}
		if result.RowsAffected == 0 {
			return product.ErrProductNotFound
		}
		return replaceCategories(tx, item.ID(), item.CategoryIDs())
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, item.ID())
}

func (r *pgRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&ProductEntity{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return product.ErrProductNotFound
	}
	return nil
}

func replaceCategories(tx *gorm.DB, productID uint64, categoryIDs []uint64) error {
	if err := tx.Where("product_id = ?", productID).Delete(&ProductCategoryEntity{}).Error; err != nil {
		return err
	}
	if len(categoryIDs) == 0 {
		return nil
	}

	relations := make([]ProductCategoryEntity, len(categoryIDs))
	for i, categoryID := range categoryIDs {
		relations[i] = ProductCategoryEntity{ProductID: productID, CategoryID: categoryID}
	}
	return tx.Create(&relations).Error
}

func loadCategoryIDs(db *gorm.DB, productIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	var relations []ProductCategoryEntity
	if err := db.Where("product_id IN ?", productIDs).
		Order("category_id ASC").
		Find(&relations).Error; err != nil {
		return nil, err
	}
	for _, relation := range relations {
		result[relation.ProductID] = append(result[relation.ProductID], relation.CategoryID)
	}
	return result, nil
}

func translateError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "idx_products_code":
			return product.ErrProductCodeAlreadyExists
		case "idx_products_sku":
			return product.ErrProductSKUAlreadyExists
		case "idx_products_barcode":
			return product.ErrProductBarcodeAlreadyExists
		}
	}
	return err
}

var _ = fmt.Sprintf
