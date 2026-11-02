package pg_product

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"gorm.io/gorm"
)

type ProductEntity struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Code        string         `gorm:"type:varchar(64);not null"`
	SKU         string         `gorm:"type:varchar(100);not null;default:''"`
	Barcode     string         `gorm:"type:varchar(100);not null;default:''"`
	Description string         `gorm:"type:text;not null;default:''"`
	IsActive    bool           `gorm:"default:true;not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (ProductEntity) TableName() string { return "products" }

type ProductCategoryEntity struct {
	ProductID  uint64 `gorm:"primaryKey;autoIncrement:false;not null;index:idx_product_categories_category_id"`
	CategoryID uint64 `gorm:"primaryKey;autoIncrement:false;not null;index:idx_product_categories_category_id"`
	CreatedAt  time.Time
}

func (ProductCategoryEntity) TableName() string { return "product_categories" }

func (e *ProductEntity) ToDomain(categoryIDs []uint64) *product.Product {
	return product.ReconstructProduct(
		e.ID, e.Name, e.Code, e.SKU, e.Barcode, e.Description,
		e.IsActive, categoryIDs, e.CreatedAt, e.UpdatedAt,
	)
}

func FromDomain(item *product.Product) *ProductEntity {
	return &ProductEntity{
		ID: item.ID(), Name: item.Name(), Code: item.Code(), SKU: item.SKU(),
		Barcode: item.Barcode(), Description: item.Description(),
		IsActive: item.IsActive(), CreatedAt: item.CreatedAt(), UpdatedAt: item.UpdatedAt(),
	}
}
