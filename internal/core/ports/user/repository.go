package user

import (
	"context"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) (*user.User, error)
	GetByID(ctx context.Context, id uint64) (*user.User, error)
	List(ctx context.Context, req paginator.Request) ([]*user.User, int64, error)
	Update(ctx context.Context, u *user.User) (*user.User, error)
	Delete(ctx context.Context, id uint64) error
}
