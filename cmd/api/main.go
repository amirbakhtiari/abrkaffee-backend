package main

import (
	"log"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/category"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/middleware"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_category"
	category2 "github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/services/auth_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/category_service"
	"github.com/amirbakhtiari/abrcoffee/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configurations: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %w", err)
	}
	r := gin.Default()
	v1 := r.Group("/api/v1")
	authMiddleware := middleware.AuthMiddleware(cfg)
	WireCategories(db, v1, authMiddleware)
	WireAuth(db, v1, cfg)

	log.Printf("🚀 AbrCoffee Server is running on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("❌ Failed to run server: %v", err)
	}
}
