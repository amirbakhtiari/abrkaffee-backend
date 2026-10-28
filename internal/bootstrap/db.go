package bootstrap

import (
	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/pkg/database"
	"gorm.io/gorm"
)

func initDB(cfg *config.Config) (*gorm.DB, error) {
	return database.NewPostgresConnection(cfg)
}
