package product

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(router *gin.RouterGroup, handler *Handler) {
	products := router.Group("/products")
	products.GET("", handler.ListPublicProducts)
	products.GET("/:id", handler.GetPublicProductByID)
}

func RegisterProtectedRoutes(router *gin.RouterGroup, handler *Handler) {
	products := router.Group("/admin/products")
	products.POST("", handler.CreateProduct)
	products.GET("", handler.ListProducts)
	products.GET("/:id", handler.GetProductByID)
	products.PUT("/:id", handler.UpdateProduct)
	products.PATCH("/:id/status", handler.UpdateProductStatus)
	products.DELETE("/:id", handler.DeleteProduct)
	products.PUT("/:id/image", handler.UpdateProductImage)
}
