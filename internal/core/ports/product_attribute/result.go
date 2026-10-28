package product_attribute

type AttributeValueResult struct {
	ID          uint64
	AttributeID uint64
	Name        string
	SortOrder   int
}

type AttributeResult struct {
	ID        uint64
	Name      string
	SortOrder int
	Values    []AttributeValueResult
}
