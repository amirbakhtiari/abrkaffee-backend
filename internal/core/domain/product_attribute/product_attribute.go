package product_attribute

import (
	"errors"
	"time"
)

type Attribute struct {
	id        uint64
	name      string
	sortOrder int
	values    []*Value
	createdAt time.Time
}

func NewAttribute(name string, sortOrder int) (*Attribute, error) {
	if name == "" {
		return nil, ErrAttributeNameRequired
	}

	if sortOrder < 0 {
		return nil, ErrInvalidSortOrder
	}

	return &Attribute{
		name:      name,
		sortOrder: sortOrder,
		values:    make([]*Value, 0),
		createdAt: time.Now(),
	}, nil
}

func (a *Attribute) AddValue(val *Value) error {
	if val == nil {
		return ErrNilValue
	}

	// ownership check (only meaningful after persistence)
	if a.id != 0 && val.attributeID != a.id {
		return errors.New("value does not belong to this attribute")
	}

	// duplicate checks
	for _, v := range a.values {
		if v.name == val.name {
			return ErrDuplicateValue
		}
		if v.sortOrder == val.sortOrder {
			return ErrDuplicateValue
		}
	}

	a.values = append(a.values, val)
	return nil
}

func (a *Attribute) Rename(name string) error {
	if name == "" {
		return ErrAttributeNameRequired
	}

	a.name = name
	return nil
}

func (a *Attribute) Update(name string, order int) error {
	if err := a.ChangeSortOrder(order); err != nil {
		return err
	}
	if err := a.Rename(name); err != nil {
		return err
	}
	return nil
}
func (a *Attribute) ChangeSortOrder(order int) error {
	if order < 0 {
		return ErrInvalidSortOrder
	}

	a.sortOrder = order
	return nil
}

func (a *Attribute) ID() uint64 {
	return a.id
}

func (a *Attribute) Name() string {
	return a.name
}

func (a *Attribute) SortOrder() int {
	return a.sortOrder
}

func (a *Attribute) Values() []*Value {
	values := make([]*Value, len(a.values))
	copy(values, a.values)
	return values
}

func (a *Attribute) CreatedAt() time.Time {
	return a.createdAt
}
