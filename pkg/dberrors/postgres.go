package dberrors

import (
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var constraintMap = map[string]error{
	"idx_categories_slug": category.ErrSlugCategoryAlreadyExists,
	"idx_categories_name": category.ErrNameCategoryAlreadyExists,
}

func Translate(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if translatedErr, exists := constraintMap[pgErr.ConstraintName]; exists {
			return translatedErr
		}
	}

	return err
}
