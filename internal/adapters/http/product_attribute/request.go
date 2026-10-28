package product_attribute

type CreateAttributeRequest struct {
	Name      string                        `json:"name"`
	SortOrder int                           `json:"sort_order"`
	Values    []CreateAttributeValueRequest `json:"values"`
}

type CreateAttributeValueRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type UpdateAttributeRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}
type ListProductAttributeRequest struct {
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
	Search string `form:"search"`
}
