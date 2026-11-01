package database

import (
	"fmt"
	"log"
	"time"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_attribute"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth/db_models"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_category"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Tehran",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&pg_user.UserModel{},
		&db_models.AccountDB{},
		&pg_category.CategoryEntity{},
		&pg_attribute.ProductAttributeModel{},
		&pg_attribute.ProductAttributeValueModel{},
	); err != nil {
		return nil, fmt.Errorf(
			"auto migrate database: %w",
			err,
		)
	}

	// اصلاح Foreign Key موجود در دیتابیس.
	if err := migrateAccountUserForeignKey(db); err != nil {
		return nil, err
	}

	if err := seedInitialAdmin(db); err != nil {
		return nil, fmt.Errorf(
			"failed to seed initial admin: %w",
			err,
		)
	}

	log.Println("Database connection established successfully.")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get underlying sql.DB: %w",
			err,
		)
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}
