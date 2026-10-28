package product_attribute

import (
	"net/http"
	"strconv"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product_attribute"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/product_attribute"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service port.ProductAttributeService
}

func NewAttributeHandler(service port.ProductAttributeService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateProductAttribute(c *gin.Context) {
	var req CreateAttributeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attr, err := domain.NewAttribute(req.Name, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range req.Values {
		val, err := domain.NewValue(0, v.Name, v.SortOrder)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = attr.AddValue(val)
		if err != nil {
			return
		}
	}

	result, err := h.service.Create(c.Request.Context(), attr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, toResponse(result))
}

func (h *Handler) GetAttributeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	attr, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(attr))
}

func (h *Handler) ListAttribute(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	q := port.ListProductAttributeQuery{
		Limit:  limit,
		Offset: offset,
		Search: c.Query("search"),
	}

	attrs, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := make([]AttributeResponse, len(attrs))
	for i, a := range attrs {
		res[i] = AttributeResponse{
			ID:        a.ID(),
			Name:      a.Name(),
			SortOrder: a.SortOrder(),
			Values:    nil,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": res,
	})
}

func (h *Handler) UpdateAttribute(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	attr, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req UpdateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = attr.Update(req.Name, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), attr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, toResponse(updated))
}

func (h *Handler) GetProductAttributeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}
	attr, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toResponse(attr))
}

func (h *Handler) DeleteProductAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}
	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{"message": "deleted"})
}
