package ports

import (
	"context"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/auth"
)

type AuthRepository interface {
	GetByUserID(ctx context.Context, userID uint64) (*auth.Account, error)
	GetByUsername(ctx context.Context, username string) (*auth.Account, error)

	CreateAccount(ctx context.Context, account *auth.Account) error
	UpdateAccount(ctx context.Context, account *auth.Account) error

	ListAccounts(ctx context.Context, limit, offset int) ([]*auth.Account, int64, error)
}
