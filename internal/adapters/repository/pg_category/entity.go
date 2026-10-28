package pg_category

import (
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"gorm.io/gorm"
)

type CategoryEntity struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(100);not null;uniqueIndex"`
	Slug      string `gorm:"type:varchar(100);not null;uniqueIndex"`
	IsActive  bool   `gorm:"default:true;not null"`
	Priority  int    `gorm:"default:0;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*CategoryEntity) TableName() string {
	return "categories"
}

func (e *CategoryEntity) ToDomain() *category.Category {
	return category.ReconstructCategory(
		e.ID,
		e.Name,
		e.Slug,
		e.Priority,
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
	)
}

func FromDomain(d *category.Category) *CategoryEntity {
	return &CategoryEntity{
		ID:        d.ID(),
		Name:      d.Name(),
		Slug:      d.Slug(),
		IsActive:  d.IsActive(),
		Priority:  d.Priority(),
		CreatedAt: d.CreatedAt(),
		UpdatedAt: d.UpdatedAt(),
	}
}
