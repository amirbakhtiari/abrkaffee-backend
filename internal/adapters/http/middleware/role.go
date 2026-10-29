package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		role, ok := roleValue.(string)

		if !exists || !ok || role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
