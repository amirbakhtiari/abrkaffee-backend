package auth_handler

import "github.com/gin-gonic/gin"

func (h *AuthHandler) RegisterAuthRoutes(
	publicRouter *gin.RouterGroup,
	adminRouter *gin.RouterGroup,
) {
	publicAuth := publicRouter.Group("/auth")
	publicAuth.POST("/login", h.LoginHandler)

	adminRouter.POST("/auth/register", h.RegisterHandler)
}
