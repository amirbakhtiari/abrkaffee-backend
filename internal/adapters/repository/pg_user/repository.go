package pg_user

import (
	"context"
	"errors"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	userport "github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) userport.UserRepository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	u *domain.User,
) (*domain.User, error) {
	if u == nil {
		return nil, domain.ErrUserIsNil
	}

	model := FromDomain(u)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return ToDomain(model), nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uint64,
) (*domain.User, error) {
	var model UserModel

	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return ToDomain(&model), nil
}

func (r *Repository) List(
	ctx context.Context,
	req paginator.Request,
) ([]*domain.User, int64, error) {
	var (
		models []UserModel
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&UserModel{})

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

func (r *Repository) Update(
	ctx context.Context,
	u *domain.User,
) (*domain.User, error) {
	if u == nil {
		return nil, domain.ErrUserIsNil
	}

	model := FromDomain(u)

	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", u.ID()).
		Updates(map[string]interface{}{
			"name":       model.Name,
			"family":     model.Family,
			"email":      model.Email,
			"mobile":     model.Mobile,
			"birth_date": model.BirthDate,
			"is_active":  model.IsActive,
			"updated_at": model.UpdatedAt,
		})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, domain.ErrUserNotFound
	}

	return ToDomain(model), nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id uint64,
) error {
	if id == 0 {
		return domain.ErrInvalidID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model UserModel

		// قفل‌کردن ردیف User برای جلوگیری از Race Condition
		// با عملیات ساخت Account.
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&model, id).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUserNotFound
		}

		if err != nil {
			return err
		}

		// بررسی وجود Account وابسته به User
		var accountCount int64

		if err := tx.
			Table("accounts").
			Where("user_id = ?", id).
			Count(&accountCount).Error; err != nil {
			return err
		}

		if accountCount > 0 {
			return domain.ErrUserHasAccount
		}

		// حذف User
		result := tx.Delete(&UserModel{}, id)

		if result.Error != nil {
			// محافظت نهایی در برابر Foreign Key Violation
			var pgErr *pgconn.PgError

			if errors.As(result.Error, &pgErr) && pgErr.Code == "23503" {
				return domain.ErrUserHasAccount
			}

			return result.Error
		}

		if result.RowsAffected == 0 {
			return domain.ErrUserNotFound
		}

		return nil
	})
}
