package bootstrap

import (
	"log"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/gateways"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler"
	httpcategory "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/category"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/middleware"
	orderhttp "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/order"
	httppricing "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/pricing"
	httpproduct "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/product"
	httpattr "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/product_attribute"
	httpuser "github.com/amirbakhtiari/abrcoffee/internal/adapters/http/user"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_attribute"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_category"
	pgorder "github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_order"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_pricing"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_product"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/storage/local"
	categorydomain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	productdomain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/services/auth_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/category_service"
	ordersvc "github.com/amirbakhtiari/abrcoffee/internal/services/order_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/pricing_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/product_attribute_service"
	"github.com/amirbakhtiari/abrcoffee/internal/services/product_service"
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

	// Initialize product image storage once.
	fileStorage, err := local.NewStorage("./uploads")
	if err != nil {
		log.Fatalf("initialize product file storage: %v", err)
	}

	log.Printf("Product image storage initialized at: %s", fileStorage.Root())

	// Serve uploaded images from the same storage directory.
	router.Static("/uploads", fileStorage.Root())

	registerAuthModule(v1, admin, db, cfg)
	registerAttributeModule(admin, db)
	registerUserModule(admin, db)
	registerCategoryModule(v1, admin, db)
	registerProductModule(v1, admin, db, fileStorage)
	registerPricingModule(admin, db)
	registerOrderModule(admin, db)

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

func registerUserModule(private *gin.RouterGroup, db *gorm.DB) {
	repo := pg_user.NewUserRepository(db)
	systemClock := clock.NewSystemClock()
	service := user_service.NewService(repo, systemClock)
	handler := httpuser.NewHandler(service)

	httpuser.RegisterUserRoutes(private, handler)

	log.Println("⚡ User module wired up successfully.")
}

func registerCategoryModule(
	v1 *gin.RouterGroup,
	admin *gin.RouterGroup,
	db *gorm.DB,
) {
	repo := pg_category.NewPostgresRepository(db)

	service := category_service.NewCategoryService(
		repo,
		categorydomain.DefaultClock,
	)

	handler := httpcategory.NewHandler(service)

	// Public: only list categories
	v1.GET("/categories", handler.ListCategories)

	// Admin: get by ID and manage categories
	httpcategory.RegisterProtectedRoutes(admin, handler)
	admin.GET("/categories/:id", handler.GetCategoryByID)

	log.Println("⚡ Category module wired up successfully.")
}

func registerProductModule(
	v1 *gin.RouterGroup,
	admin *gin.RouterGroup,
	db *gorm.DB,
	fileStorage *local.Storage,
) {
	productRepo := pg_product.NewPostgresRepository(db)
	categoryRepo := pg_category.NewPostgresRepository(db)

	service := product_service.NewProductService(
		productRepo,
		categoryRepo,
		fileStorage,
		productdomain.DefaultClock,
	)

	handler := httpproduct.NewHandler(service)

	httpproduct.RegisterPublicRoutes(v1, handler)
	httpproduct.RegisterProtectedRoutes(admin, handler)

	log.Println("⚡ Product module wired up successfully.")
}

func registerPricingModule(admin *gin.RouterGroup, db *gorm.DB) {
	repo := pg_pricing.NewPostgresRepository(db)
	service := pricing_service.NewPricingService(repo)
	handler := httppricing.NewHandler(service)

	httppricing.RegisterProtectedRoutes(admin, handler)

	log.Println("⚡ Pricing module wired up successfully.")
}

func registerOrderModule(v1 *gin.RouterGroup, db *gorm.DB) {
	productRepo := pg_product.NewPostgresRepository(db)
	categoryRepo := pg_category.NewPostgresRepository(db)
	pricingRepo := pg_pricing.NewPostgresRepository(db)

	productCatalog := gateways.NewProductCatalogGateway(productRepo, categoryRepo)
	pricingCatalog := gateways.NewPricingCatalogGateway(pricingRepo)

	orderRepo := pgorder.NewPostgresRepository(db)
	orderService := ordersvc.NewOrderService(orderRepo, productCatalog, pricingCatalog)

	orderHandler := orderhttp.NewHandler(orderService)
	orderhttp.RegisterRoutes(v1, orderHandler)

	log.Println("⚡ Order module wired up successfully.")
}
