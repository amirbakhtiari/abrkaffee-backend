package user

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidID      = errors.New("invalid id")
	ErrUserIsNil      = errors.New("user is nil")
	ErrUserHasAccount = errors.New("user has an account and cannot be deleted")
	ErrUserInactive   = errors.New("user is inactive")
)
