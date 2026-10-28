package product_attribute

type CreateAttributeValueDTO struct {
	Name      string
	SortOrder int
}

type CreateAttributeDTO struct {
	Name      string
	SortOrder int
	Values    []CreateAttributeValueDTO
}

type UpdateAttributeValueDTO struct {
	ID        uint64
	Name      string
	SortOrder int
}

type UpdateAttributeDTO struct {
	Name      string
	SortOrder int
	Values    []UpdateAttributeValueDTO
}
