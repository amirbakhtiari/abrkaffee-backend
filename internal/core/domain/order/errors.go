package order

import "errors"

var (
	ErrOrderNotFound             = errors.New("order not found")
	ErrEmptyOrderItems           = errors.New("order must contain at least one item")
	ErrInvalidQuantity           = errors.New("item quantity must be greater than zero")
	ErrInvalidUnitPrice          = errors.New("unit price cannot be negative")
	ErrInvalidTableID            = errors.New("invalid table ID")
	ErrInvalidStatusTransition   = errors.New("invalid status transition")
	ErrOrderAlreadyCompleted     = errors.New("completed order cannot be altered")
	ErrCannotCompleteUnpaidOrder = errors.New("order cannot be completed until fully paid and served")
	ErrInvalidPaymentAmount      = errors.New("payment amount must be greater than zero")
	ErrPaymentExceedsRemaining   = errors.New("payment amount exceeds remaining unpaid balance")
	ErrProductNotActive          = errors.New("one or more products are inactive")
	ErrActivePriceNotFound       = errors.New("active approved price not found for product")
	ErrOptimisticLockConflict    = errors.New("order was updated by another transaction")
)
