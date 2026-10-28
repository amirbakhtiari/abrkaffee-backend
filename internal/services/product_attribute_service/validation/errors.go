package validation

type Error struct {
	Field   string
	Message string
	Value   any
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(field, message string, value any) *Error {
	return &Error{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

type Errors []*Error

func (e Errors) Error() string {
	if len(e) == 0 {
		return "validation error"
	}
	return e[0].Message
}

func Is(err error) bool {
	_, ok1 := err.(*Error)
	_, ok2 := err.(Errors)
	return ok1 || ok2
}
