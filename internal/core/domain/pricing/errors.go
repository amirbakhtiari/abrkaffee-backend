package pricing

import "errors"

var (
	ErrAnnouncementNotFound       = errors.New("price announcement not found")
	ErrAnnouncementTitleRequired  = errors.New("announcement title is required")
	ErrAnnouncementCodeRequired   = errors.New("announcement code is required")
	ErrAnnouncementCodeDuplicate  = errors.New("announcement code already exists")
	ErrAnnouncementAlreadyHandled = errors.New("announcement is already approved or rejected")
	ErrInvalidItemPrice           = errors.New("item price must be greater than zero")
	ErrInvalidDiscountPercent     = errors.New("discount percent must be between 0 and 100")
	ErrNoItemsProvided            = errors.New("announcement must contain at least one item")
	ErrDuplicateProductInItems    = errors.New("duplicate product in announcement items")
	ErrProductNotFound            = errors.New("one or more specified products do not exist")
)
