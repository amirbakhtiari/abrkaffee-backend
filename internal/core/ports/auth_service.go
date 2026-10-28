package ports

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type RegisterResult struct {
	User  *user.User
	Token string
}

type AuthService interface {
	Register(ctx context.Context, name, family, username, password, mobile string) (*user.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	UpdateProfile(ctx context.Context, userID uint64, name, family, mobile string, birthDate *time.Time) error
	UpdatePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error

	ChangeStatus(ctx context.Context, userID uint64, targetStatus bool) error
	GetUsersList(ctx context.Context, page, pageSize int) ([]*user.User, paginator.Pagination, error)
}
