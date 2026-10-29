package presenter

import (
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type UserResponse struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Family   string `json:"family"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	IsActive bool   `json:"is_active"`
}
type ListUserResponse struct {
	Pagination paginator.Meta `json:"pagination"`
	Data       []UserResponse `json:"data"`
}

func FromDomain(u *user.User) UserResponse {
	return UserResponse{
		ID:       u.ID(),
		Name:     u.Name(),
		Family:   u.Family(),
		Email:    u.Email(),
		Mobile:   u.Mobile(),
		IsActive: u.IsActive(),
	}
}

func FromDomainList(users []*user.User) []UserResponse {
	list := make([]UserResponse, len(users))
	for i, u := range users {
		list[i] = FromDomain(u)
	}
	return list
}
