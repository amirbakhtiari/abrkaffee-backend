package category

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/category"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ports.CategoryService
}

func NewHandler(service ports.CategoryService) *Handler {
	return &Handler{service: service}
}
func (h *Handler) CreateCategory(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dto := &ports.CreateCategoryDTO{
		Name:     req.Name,
		Slug:     req.Slug,
		Priority: req.Priority,
		IsActive: req.IsActive,
	}

	result, err := h.service.CreateCategory(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": ToResponse(result)})
}
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto := &ports.UpdateCategoryDTO{
		Name:     req.Name,
		Slug:     req.Slug,
		Priority: req.Priority,
		IsActive: req.IsActive,
	}

	result, err := h.service.UpdateCategory(c.Request.Context(), id, dto)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(result))
}
func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "category deleted successfully"})
}
func (h *Handler) GetCategoryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	result, err := h.service.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": ToResponse(result)})
}
func (h *Handler) ListCategories(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto := ports.ListCategoriesDTO{
		Skip: query.Skip,
		Take: query.Take,
	}

	results, err := h.service.ListCategories(c.Request.Context(), dto)

	responseList := make([]CategoryResponse, len(results))
	for i, result := range results {
		responseList[i] = ToResponse(result)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, responseList)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, category.ErrCategoryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, category.ErrCategoryNameEmpty),
		errors.Is(err, category.ErrCategoryNameTooLong),
		errors.Is(err, category.ErrCategorySlugInvalid),
		errors.Is(err, category.ErrCategoryPriorityInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
