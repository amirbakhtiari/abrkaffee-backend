package pg_auth

import (
	"context"
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth/db_models"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/auth"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"gorm.io/gorm"
)

var _ ports.AuthRepository = (*PostgresAuthRepository)(nil)

type PostgresAuthRepository struct {
	db *gorm.DB
}

func NewPostgresAuthRepository(db *gorm.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

func (repo *PostgresAuthRepository) GetByUserID(
	ctx context.Context,
	userID uint64,
) (*auth.Account, error) {
	var model db_models.AccountDB

	err := repo.db.WithContext(ctx).
		Preload("User").
		First(&model, "user_id = ?", userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return repo.toDomain(&model), nil
}

func (repo *PostgresAuthRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*auth.Account, error) {
	var model db_models.AccountDB

	err := repo.db.WithContext(ctx).
		Preload("User").
		First(&model, "username = ?", username).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return repo.toDomain(&model), nil
}

func (repo *PostgresAuthRepository) CreateAccount(
	ctx context.Context,
	account *auth.Account,
) error {
	if account == nil || account.Profile == nil {
		return errors.New("account and profile are required")
	}

	model := &db_models.AccountDB{
		UserID:       account.UserID,
		Username:     account.Username,
		PasswordHash: account.PasswordHash,
		Role:         account.Role,
		IsActive:     account.IsActive,
		CreatedAt:    account.CreatedAt,
		UpdatedAt:    account.UpdatedAt,
	}

	return repo.db.WithContext(ctx).Create(model).Error
}

func (repo *PostgresAuthRepository) UpdateAccount(
	ctx context.Context,
	account *auth.Account,
) error {
	if account == nil {
		return errors.New("account is required")
	}

	return repo.db.WithContext(ctx).
		Model(&db_models.AccountDB{}).
		Where("user_id = ?", account.UserID).
		Updates(map[string]interface{}{
			"username":      account.Username,
			"password_hash": account.PasswordHash,
			"is_active":     account.IsActive,
			"updated_at":    account.UpdatedAt,
		}).Error
}

func (repo *PostgresAuthRepository) ListAccounts(
	ctx context.Context,
	limit int,
	offset int,
) ([]*auth.Account, int64, error) {
	var models []db_models.AccountDB
	var total int64

	query := repo.db.WithContext(ctx).Model(&db_models.AccountDB{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := repo.db.WithContext(ctx).
		Preload("User").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	accounts := make([]*auth.Account, 0, len(models))
	for i := range models {
		accounts = append(accounts, repo.toDomain(&models[i]))
	}

	return accounts, total, nil
}

func (repo *PostgresAuthRepository) toDomain(
	model *db_models.AccountDB,
) *auth.Account {
	profile := user.Rehydrate(
		model.User.ID,
		model.User.Name,
		model.User.Family,
		model.User.Email,
		model.User.Mobile,
		nil,
		model.User.IsActive,
		model.User.CreatedAt,
		model.User.UpdatedAt,
	)

	return &auth.Account{
		ID:           model.ID,
		UserID:       model.UserID,
		Profile:      profile,
		Username:     model.Username,
		PasswordHash: model.PasswordHash,
		Role:         model.Role,
		IsActive:     model.IsActive,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
