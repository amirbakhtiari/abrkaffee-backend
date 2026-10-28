package auth_handler

import (
	"net/http"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler/dto"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler/presenter"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
)
import "github.com/gin-gonic/gin"

type AuthHandler struct {
	svc ports.AuthService
}

func NewAuthHandler(svc ports.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.Register(c.Request.Context(),
		req.Name,
		req.Family,
		req.Username,
		req.Password,
		req.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := presenter.FromDomain(user)
	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"data":    response,
	})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
		"token":   token,
	})
}
