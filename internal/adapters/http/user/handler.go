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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid birth_date",
			})
			return
		}
		birthDate = &t
	}

	result, err := h.service.Create(
		c.Request.Context(),
		req.Name,
		req.Family,
		req.Mobile,
		req.Email,
		birthDate,
	)
	if err != nil {
		if errors.Is(err, userdomain.ErrInvalidNameRequired) ||
			errors.Is(err, userdomain.ErrInvalidFamilyRequired) ||
			errors.Is(err, userdomain.ErrInvalidEmail) ||
			errors.Is(err, userdomain.ErrInvalidMobileRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
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
	result, meta, err := h.service.List(c.Request.Context(), paginator.Request{
		Page:     int(page),
		PageSize: int(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]UserResponse, 0, len(result))
	for _, u := range result {
		if u == nil {
			continue
		}
		items = append(items, toResponse(u))
	}

	c.JSON(http.StatusOK, ListUsersResponse{
		Items: items,
		Meta:  paginator.NewMeta(int(page), int(pageSize), meta),
	})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	// 1. Parse user ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, userdomain.ErrInvalidID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// 2. Parse request body
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// 3. Parse birth date
	var birthDate *time.Time

	if strings.TrimSpace(req.BirthDate) != "" {
		parsed, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "birth_date must use YYYY-MM-DD format",
			})
			return
		}

		birthDate = &parsed
	}

	// 4. Build service DTO
	input := user.UpdateUserDTO{
		Name:      req.Name,
		Family:    req.Family,
		Mobile:    req.Mobile,
		Email:     req.Email,
		BirthDate: birthDate,
	}

	// 5. Call service
	result, err := h.service.Update(
		c.Request.Context(),
		id,
		input,
	)
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

	// 6. Return response
	c.JSON(http.StatusOK, toResponse(result))
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, userdomain.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, userdomain.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
