package pg_attribute

import "time"

type ProductAttributeModel struct {
	ID        uint64                       `gorm:"primaryKey;autoIncrement"`
	Name      string                       `gorm:"type:varchar(255);not null"`
	SortOrder int                          `gorm:"default:0"`
	Values    []ProductAttributeValueModel `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time                    `gorm:"default:NOW()"`
}

func (ProductAttributeModel) TableName() string {
	return "product_attributes"
}

type ProductAttributeValueModel struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	AttributeID uint64    `gorm:"not null;index"`
	Name        string    `gorm:"type:varchar(255);not null"`
	SortOrder   int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"default:NOW()"`
}

func (ProductAttributeValueModel) TableName() string {
	return "product_attribute_values"
}
