package category

type CreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Slug     string `json:"slug" binding:"required"`
	Priority int    `json:"priority" binding:"min=0"`
	IsActive bool   `json:"is_active"`
}

type UpdateRequest struct {
	Name     string `json:"name" binding:"required"`
	Slug     string `json:"slug" binding:"required"`
	Priority int    `json:"priority" binding:"min=0"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

type ListQuery struct {
	Skip uint64 `form:"skip,default=0"`
	Take uint64 `form:"take,default=10"`
}
