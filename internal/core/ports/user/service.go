package user

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type Service interface {
	Create(
		ctx context.Context,
		name, family, mobile, email string,
		birthDate *time.Time,
	) (*user.User, error)

	GetByID(ctx context.Context, id uint64) (*user.User, error)

	List(
		ctx context.Context,
		req paginator.Request,
	) ([]*user.User, int64, error)

	Update(
		ctx context.Context,
		id uint64,
		dto UpdateUserDTO,
	) (*user.User, error)

	Delete(ctx context.Context, id uint64) error

	SetActive(
		ctx context.Context,
		id uint64,
		isActive bool,
	) (*user.User, error)
}
