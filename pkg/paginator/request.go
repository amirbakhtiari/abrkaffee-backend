package paginator

type Request struct {
	Page     int
	PageSize int
}

func NewRequest(page, pageSize int) Request {
	r := Request{
		Page:     page,
		PageSize: pageSize,
	}
	r.Normalize()
	return r
}

func (r *Request) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}

	if r.PageSize < 1 {
		r.PageSize = 10
	}

	if r.PageSize > 100 {
		r.PageSize = 100
	}
}

func (r Request) Offset() int {
	return (r.Page - 1) * r.PageSize
}
