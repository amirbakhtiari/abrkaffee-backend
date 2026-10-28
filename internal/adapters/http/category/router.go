package category

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(router *gin.RouterGroup, handler *Handler) {
	categories := router.Group("/categories")
	{
		categories.GET("", handler.ListCategories)
		categories.GET("/:id", handler.GetCategoryByID)
	}
}
func RegisterProtectedRoutes(router *gin.RouterGroup, handler *Handler) {
	categories := router.Group("/categories")
	{
		categories.POST("", handler.CreateCategory)
		categories.PUT("/:id", handler.UpdateCategory)
		categories.DELETE("/:id", handler.DeleteCategory)
	}
}
