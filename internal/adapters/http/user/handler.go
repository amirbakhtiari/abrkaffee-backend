package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	userdomain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service user.Service
}

func NewHandler(service user.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var birthDate *time.Time
	if req.BirthDate != "" {
		t, err := time.Parse(time.DateOnly, req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date"})
			return
		}
		birthDate = &t
	}

	result, err := h.service.Create(c.Request.Context(), req.Name, req.Family, req.Mobile, req.Email, birthDate)
	if err != nil {
		if errors.Is(err, userdomain.ErrInvalidNameRequired) ||
			errors.Is(err, userdomain.ErrInvalidFamilyRequired) ||
			errors.Is(err, userdomain.ErrInvalidEmail) ||
			errors.Is(err, userdomain.ErrInvalidMobileRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, toResponse(result))
}

func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toResponse(result))
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, err := strconv.ParseUint(c.Query("page"), 10, 64)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}
	pageSize, err := strconv.ParseUint(c.Query("page_size"), 10, 64)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size"})
		return
	}
	result, meta, err := h.service.List(c.Request.Context(), paginator.Request{Page: int(page), PageSize: int(pageSize)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]UserResponse, 0, len(result))
	for _, u := range result {
		if u != nil {
			items = append(items, toResponse(u))
		}
	}
	c.JSON(http.StatusOK, ListUsersResponse{
		Items: items,
		Meta:  paginator.NewMeta(int(page), int(pageSize), meta),
	})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	var birthDate *time.Time
	if strings.TrimSpace(req.BirthDate) != "" {
		parsed, err := time.Parse(time.DateOnly, req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "birth_date must use YYYY-MM-DD format"})
			return
		}
		birthDate = &parsed
	}
	result, err := h.service.Update(c.Request.Context(), id, user.UpdateUserDTO{
		Name: req.Name, Family: req.Family, Mobile: req.Mobile, Email: req.Email, BirthDate: birthDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, userdomain.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		case errors.Is(err, userdomain.ErrInvalidNameRequired),
			errors.Is(err, userdomain.ErrInvalidFamilyRequired),
			errors.Is(err, userdomain.ErrInvalidEmail),
			errors.Is(err, userdomain.ErrInvalidMobileRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		}
		return
	}
	c.JSON(http.StatusOK, toResponse(result))
}

func (h *Handler) SetUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req SetUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_active is required and must be a boolean"})
		return
	}
	if err := h.service.SetActive(c.Request.Context(), id, *req.IsActive); err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, userdomain.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, userdomain.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		case errors.Is(err, userdomain.ErrUserHasOperationalHistory):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
