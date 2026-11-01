package category

import "errors"

var (
	ErrCategoryNotFound          = errors.New("category not found")
	ErrCategoryNameEmpty         = errors.New("category name cannot be empty")
	ErrCategoryNameTooLong       = errors.New("category name is too long (max 100 chars)")
	ErrCategorySlugInvalid       = errors.New("category slug is invalid")
	ErrCategoryPriorityInvalid   = errors.New("category priority cannot be negative")
	ErrSlugCategoryAlreadyExists = errors.New("slug وارد شده تکراری هست")
	ErrNameCategoryAlreadyExists = errors.New("نام وارد شده تکراری است")
	ErrCategoryRequired          = errors.New("category is required")
)
