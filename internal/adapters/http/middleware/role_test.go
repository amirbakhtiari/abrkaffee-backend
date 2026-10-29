package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		setRole        bool
		expectedStatus int
	}{
		{
			name:           "admin is allowed",
			role:           "Admin",
			setRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "staff is forbidden",
			role:           "Staff",
			setRole:        true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing role is forbidden",
			setRole:        false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.GET("/admin", func(c *gin.Context) {
				if tt.setRole {
					c.Set("role", tt.role)
				}
				RequireRole("Admin")(c)
				if c.IsAborted() {
					return
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}
		})
	}
}
