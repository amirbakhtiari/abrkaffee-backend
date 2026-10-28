package product_attribute

import "errors"

var (
	ErrAttributeNameRequired = errors.New("attribute name is required")
	ErrValueNameRequired     = errors.New("value name is required")
	ErrDuplicateValue        = errors.New("duplicate value")
	ErrInvalidSortOrder      = errors.New("invalid sort order")
	ErrNilValue              = errors.New("nil value")
)
