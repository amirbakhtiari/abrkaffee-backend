package product

import "errors"

var (
	ErrProductNotFound             = errors.New("product not found")
	ErrProductRequired             = errors.New("product is required")
	ErrProductNameRequired         = errors.New("product name is required")
	ErrProductNameTooLong          = errors.New("product name is too long")
	ErrProductCodeRequired         = errors.New("product code is required")
	ErrProductCodeTooLong          = errors.New("product code is too long")
	ErrProductSKUTooLong           = errors.New("product SKU is too long")
	ErrProductBarcodeTooLong       = errors.New("product barcode is too long")
	ErrProductDescriptionTooLong   = errors.New("product description is too long")
	ErrProductCategoryRequired     = errors.New("at least one category is required")
	ErrProductCategoryInvalid      = errors.New("one or more categories do not exist")
	ErrProductCodeAlreadyExists    = errors.New("product code already exists")
	ErrProductSKUAlreadyExists     = errors.New("product SKU already exists")
	ErrProductBarcodeAlreadyExists = errors.New("product barcode already exists")
	ErrProductImageRequired        = errors.New("product image is required")
)
