package user

import (
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type UserResponse struct {
	ID        uint64     `json:"id"`
	Name      string     `json:"name"`
	Family    string     `json:"family"`
	Email     string     `json:"email"`
	Mobile    string     `json:"mobile"`
	IsActive  bool       `json:"is_active"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func toResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID(),
		Name:      u.Name(),
		Family:    u.Family(),
		Email:     u.Email(),
		Mobile:    u.Mobile(),
		IsActive:  u.IsActive(),
		BirthDate: u.BirthDate(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}

type ListUsersResponse struct {
	Items []UserResponse `json:"items"`
	Meta  paginator.Meta `json:"meta"`
}
