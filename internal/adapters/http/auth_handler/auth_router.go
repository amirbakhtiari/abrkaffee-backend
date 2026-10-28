package auth_handler

import "github.com/gin-gonic/gin"

func (h *AuthHandler) RegisterAuthRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", h.RegisterHandler)
		authGroup.POST("/login", h.LoginHandler)
	}
}
