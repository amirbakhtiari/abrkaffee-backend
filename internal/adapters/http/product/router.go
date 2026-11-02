package product

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(router *gin.RouterGroup, handler *Handler) {
	products := router.Group("/products")
	products.GET("", handler.ListProducts)
	products.GET("/:id", handler.GetProductByID)
}

func RegisterProtectedRoutes(router *gin.RouterGroup, handler *Handler) {
	products := router.Group("/products")
	products.POST("", handler.CreateProduct)
	products.PUT("/:id", handler.UpdateProduct)
	products.DELETE("/:id", handler.DeleteProduct)
}
