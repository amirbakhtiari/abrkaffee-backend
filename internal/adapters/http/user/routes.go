package user

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	users := router.Group("/users")

	{
		users.POST("", handler.CreateUser)
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUserByID)
		users.PUT("/:id", handler.UpdateUser)

		users.DELETE("/:id", handler.DeleteUser)

		users.PATCH("/:id/status", handler.SetUserActive)
	}
}
