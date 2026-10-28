package pg_user

import (
	"context"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	model := FromDomain(u)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return ToDomain(&model), nil
}

func (r *Repository) List(ctx context.Context, req paginator.Request) ([]*domain.User, int64, error) {
	var (
		models []UserModel
		total  int64
	)
	query := r.db.WithContext(ctx).Model(&models)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(req.PageSize).
		Offset(req.Offset()).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = ToDomain(&models[i])
	}
	return users, total, nil
}

func (r *Repository) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	model := FromDomain(u)

	if err := r.db.WithContext(ctx).
		Save(model).Error; err != nil {
		return nil, err
	}

	return ToDomain(model), nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, id).Error
}
