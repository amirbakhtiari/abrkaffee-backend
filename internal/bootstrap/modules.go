package bootstrap

import (
	"log"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler"
	httpcategory "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/category"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/middleware"
	httpattr "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/product_attribute"
	httpuser "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/user"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_attribute"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_category"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
	categorydomain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/services/auth_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/category_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/product_attribute_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/user_service"
	"github.com/amirbakhtiari/abrcoffee/internal/shared/clock"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerMiddleware(router *gin.Engine, cfg *config.Config) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
}

func registerModules(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	v1 := router.Group("/api/v1")

	auth := middleware.AuthMiddleware(cfg)

	// Routes requiring authentication
	private := v1.Group("")
	private.Use(auth)

	// Routes requiring Admin role
	admin := v1.Group("")
	admin.Use(auth, middleware.RequireRole("Admin"))

	registerAuthModule(v1, admin, db, cfg)
	registerAttributeModule(admin, db)
	registerUserModule(admin, db)
	registerCategoryModule(v1, private, db)
}

func registerAttributeModule(v1 *gin.RouterGroup, db *gorm.DB) {

	repo := pg_attribute.NewAttributeRepository(db)
	service := product_attribute_service.NewProductAttributeService(repo)
	handler := httpattr.NewAttributeHandler(service)

	attr := v1.Group("/attributes")
	{
		attr.POST("", handler.CreateProductAttribute)
		attr.GET("", handler.ListAttribute)
		attr.GET("/:id", handler.GetProductAttributeByID)
		attr.PUT("/:id", handler.UpdateAttribute)
		attr.DELETE("/:id", handler.DeleteProductAttribute)
	}

	log.Println("⚡ Attribute module wired up successfully.")
}

func registerAuthModule(
	v1 *gin.RouterGroup,
	admin *gin.RouterGroup,
	db *gorm.DB,
	cfg *config.Config,
) {
	authRepo := pg_auth.NewPostgresAuthRepository(db)
	userRepo := pg_user.NewUserRepository(db)

	svc := auth_service.NewAuthService(authRepo, userRepo, cfg)
	handler := auth_handler.NewAuthHandler(svc)

	handler.RegisterAuthRoutes(v1, admin)
	log.Println("⚡ Auth module wired up successfully.")
}

func registerCategoryModule(v1 *gin.RouterGroup, private *gin.RouterGroup, db *gorm.DB) {
	repo := pg_category.NewPostgresRepository(db)
	service := category_service.NewCategoryService(repo, categorydomain.DefaultClock)
	handler := httpcategory.NewHandler(service)
	httpcategory.RegisterProtectedRoutes(private, handler)
	httpcategory.RegisterPublicRoutes(v1, handler)
	log.Println("⚡ Category module wired up successfully.")
}

func registerUserModule(private *gin.RouterGroup, db *gorm.DB) {
	repo := pg_user.NewUserRepository(db)
	systemClock := clock.NewSystemClock()
	service := user_service.NewService(repo, systemClock)
	handler := httpuser.NewHandler(service)

	httpuser.RegisterUserRoutes(private, handler)

	log.Println("⚡ User module wired up successfully.")
}
