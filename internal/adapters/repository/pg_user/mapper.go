package pg_user

import "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"

func ToDomain(m *UserModel) *user.User {
	if m == nil {
		return nil
	}

	u := user.Rehydrate(
		m.ID,
		m.Name,
		m.Family,
		m.Email,
		m.Mobile,
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	)

	return u
}

func FromDomain(u *user.User) *UserModel {
	if u == nil {
		return nil
	}

	return &UserModel{
		ID:        u.ID(),
		Name:      u.Name(),
		Family:    u.Family(),
		Email:     u.Email(),
		Mobile:    u.Mobile(),
		IsActive:  u.IsActive(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}
