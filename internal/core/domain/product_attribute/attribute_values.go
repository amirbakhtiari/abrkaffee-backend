package product_attribute

import (
	"errors"
	"time"
)

type Value struct {
	id          uint64
	attributeID uint64
	name        string
	sortOrder   int
	createdAt   time.Time
}

func NewValue(attributeID uint64, name string, sortOrder int) (*Value, error) {
	if attributeID == 0 {
		return nil, errors.New("attributeID is required")
	}

	if name == "" {
		return nil, ErrValueNameRequired
	}

	if sortOrder < 0 {
		return nil, ErrInvalidSortOrder
	}

	return &Value{
		attributeID: attributeID,
		name:        name,
		sortOrder:   sortOrder,
		createdAt:   time.Now(),
	}, nil
}

func (v *Value) ID() uint64 {
	return v.id
}

func (v *Value) AttributeID() uint64 {
	return v.attributeID
}

func (v *Value) Name() string {
	return v.name
}

func (v *Value) SortOrder() int {
	return v.sortOrder
}

func (v *Value) CreatedAt() time.Time {
	return v.createdAt
}
