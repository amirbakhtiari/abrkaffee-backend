package database

import (
	"fmt"
	"log"
	"strings"
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
		return nil, fmt.Errorf("auto migrate database: %w", err)
	}

	if err := migrateCategoryIndexes(db); err != nil {
		return nil, fmt.Errorf("migrate category indexes: %w", err)
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

	log.Println("Database connection established successfully.")

	return db, nil
}

func migrateCategoryIndexes(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Prevent concurrent application instances from running
		// this migration simultaneously.
		if err := tx.Exec(
			"SELECT pg_advisory_xact_lock(72419031)",
		).Error; err != nil {
			return fmt.Errorf("lock category index migration: %w", err)
		}

		indexes := []struct {
			name   string
			column string
		}{
			{name: "idx_categories_name", column: "name"},
			{name: "idx_categories_slug", column: "slug"},
		}

		for _, index := range indexes {
			var existing struct {
				TableName string `gorm:"column:tablename"`
				IndexDef  string `gorm:"column:indexdef"`
			}

			result := tx.Raw(`
				SELECT tablename, indexdef
				FROM pg_indexes
				WHERE schemaname = current_schema()
				  AND indexname = ?
			`, index.name).Scan(&existing)

			if result.Error != nil {
				return fmt.Errorf("inspect index %s: %w", index.name, result.Error)
			}

			if result.RowsAffected > 0 {
				if existing.TableName != "categories" {
					return fmt.Errorf(
						"index %s belongs to unexpected table %s",
						index.name,
						existing.TableName,
					)
				}

				definition := strings.ToLower(
					strings.Join(strings.Fields(existing.IndexDef), " "),
				)

				isCorrect := strings.Contains(definition, "create unique index") &&
					strings.Contains(definition, "using btree ("+index.column+")") &&
					strings.Contains(definition, "where (deleted_at is null)")

				if isCorrect {
					continue
				}

				if err := tx.Exec(
					"DROP INDEX IF EXISTS " + index.name,
				).Error; err != nil {
					return fmt.Errorf("drop index %s: %w", index.name, err)
				}
			}

			query := fmt.Sprintf(`
				CREATE UNIQUE INDEX %s
				ON categories (%s)
				WHERE deleted_at IS NULL
			`, index.name, index.column)

			if err := tx.Exec(query).Error; err != nil {
				return fmt.Errorf("create partial index %s: %w", index.name, err)
			}
		}

		return nil
	})
}
