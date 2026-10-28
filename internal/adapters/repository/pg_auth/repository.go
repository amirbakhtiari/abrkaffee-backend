package pg_auth

import (
	"context"
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth/db_models"
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

func (repo *PostgresAuthRepository) GetByID(ctx context.Context, id uint64) (*user.User, error) {
	var dbModel db_models.UserDB
	err := repo.db.WithContext(ctx).First(&dbModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbModel.ToDomain(), nil
}

func (repo *PostgresAuthRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var dbModel db_models.UserDB
	err := repo.db.WithContext(ctx).Where("username=?", username).First(&dbModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbModel.ToDomain(), nil
}

func (repo *PostgresAuthRepository) GetByMobile(ctx context.Context, mobile string) (*user.User, error) {
	var dbModel db_models.UserDB
	err := repo.db.WithContext(ctx).Where("mobile=?", mobile).First(&dbModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbModel.ToDomain(), nil
}

func (repo *PostgresAuthRepository) UpdateUser(ctx context.Context, user *user.User) error {
	dbModel := db_models.FromDomain(user)
	return repo.db.WithContext(ctx).Save(dbModel).Error
}

func (repo *PostgresAuthRepository) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, int64, error) {
	var dbModels []db_models.UserDB
	var total int64

	err := repo.db.WithContext(ctx).Model(&dbModels).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = repo.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("id desc").
		Find(&dbModels).Error
	if err != nil {
		return nil, 0, err
	}
	users := make([]*user.User, len(dbModels))
	for i, dbModel := range dbModels {
		users[i] = dbModel.ToDomain()
	}
	return users, total, nil
}

func (repo *PostgresAuthRepository) CreateUser(ctx context.Context, user *user.User) error {
	dbModel := db_models.FromDomain(user)
	if err := repo.db.WithContext(ctx).Create(dbModel).Error; err != nil {
		return err
	}
	user.ID = dbModel.ID
	return nil
}
