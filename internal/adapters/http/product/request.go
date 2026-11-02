package product

type CreateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Code        string   `json:"code" binding:"required"`
	SKU         string   `json:"sku"`
	Barcode     string   `json:"barcode"`
	Description string   `json:"description"`
	IsActive    *bool    `json:"is_active"`
	CategoryIDs []uint64 `json:"category_ids"`
}

type UpdateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Code        string   `json:"code" binding:"required"`
	SKU         string   `json:"sku"`
	Barcode     string   `json:"barcode"`
	Description string   `json:"description"`
	IsActive    *bool    `json:"is_active" binding:"required"`
	CategoryIDs []uint64 `json:"category_ids"`
}

type ListQuery struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	PageSize   int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search     string `form:"search"`
	IsActive   *bool  `form:"is_active"`
	CategoryID uint64 `form:"category_id"`
}
