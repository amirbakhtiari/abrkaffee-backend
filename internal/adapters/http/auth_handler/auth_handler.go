package auth_handler

import (
	"errors"
	"net/http"

	"github.com/amirbakhtiari/abrcoffee/internal/adapters/http/auth_handler/dto"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/amirbakhtiari/abrcoffee/internal/services/auth_service"
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.svc.Register(
		c.Request.Context(),
		req.UserID,
		req.Username,
		req.Password,
	)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, auth_service.ErrUserNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, auth_service.ErrUserAlreadyExists) {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "account created successfully",
		"data": gin.H{
			"id":        account.ID,
			"user_id":   account.UserID,
			"username":  account.Username,
			"is_active": account.IsActive,
		},
	})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
