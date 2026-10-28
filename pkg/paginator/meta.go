package paginator

import "math"

type Meta struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"pageSize"`
	TotalRow  int64 `json:"totalRow"`
	TotalPage int   `json:"totalPage"`
}

func NewMeta(page, pageSize int, total int64) Meta {
	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))

	if totalPage <= 0 {
		totalPage = 1
	}

	return Meta{
		Page:      page,
		PageSize:  pageSize,
		TotalRow:  total,
		TotalPage: totalPage,
	}
}
