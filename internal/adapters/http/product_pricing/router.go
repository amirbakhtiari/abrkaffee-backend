package product_pricing

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	products := router.Group("/products")
	products.GET("/with-prices", handler.ListPublicProductsWithPrices)
}

func RegisterProtectedRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	products := router.Group("/admin/products")
	products.GET("/with-prices", handler.ListProductsWithPrices)
}
