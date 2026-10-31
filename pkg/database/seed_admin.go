package database

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth/db_models"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
	"github.com/amirbakhtiari/abrcoffee/pkg/utils"
	"gorm.io/gorm"
)

func seedInitialAdmin(db *gorm.DB) error {
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	family := strings.TrimSpace(os.Getenv("ADMIN_FAMILY"))
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	mobile := strings.TrimSpace(os.Getenv("ADMIN_MOBILE"))
	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	password := os.Getenv("ADMIN_PASSWORD")

	if name == "" || family == "" || email == "" ||
		mobile == "" || username == "" || password == "" {
		return errors.New("initial admin configuration is incomplete")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var existing db_models.AccountDB

		err := tx.Where("username = ?", username).First(&existing).Error
		if err == nil {
			if existing.Role != "Admin" {
				return fmt.Errorf(
					"configured admin username already belongs to a non-admin account",
				)
			}
			// The initial admin already exists. Do not recreate or overwrite it.
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check initial admin: %w", err)
		}

		passwordHash, err := utils.NewPasswordHash(password)
		if err != nil {
			return fmt.Errorf("hash initial admin password: %w", err)
		}

		now := time.Now()

		profile := &pg_user.UserModel{
			Name:      name,
			Family:    family,
			Email:     email,
			Mobile:    mobile,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := tx.Create(profile).Error; err != nil {
			return fmt.Errorf("create initial admin profile: %w", err)
		}

		account := &db_models.AccountDB{
			UserID:       profile.ID,
			Username:     username,
			PasswordHash: passwordHash,
			Role:         "Admin",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := tx.Create(account).Error; err != nil {
			return fmt.Errorf("create initial admin account: %w", err)
		}

		return nil
	})
}
