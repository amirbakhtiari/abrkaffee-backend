package product

import (
	"errors"
	"net/http"
	"strconv"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/product"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ports.ProductService
}

func NewHandler(service ports.ProductService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	result, err := h.service.CreateProduct(c.Request.Context(), &ports.CreateProductDTO{
		Name: req.Name, Code: req.Code, SKU: req.SKU, Barcode: req.Barcode,
		Description: req.Description, IsActive: isActive, CategoryIDs: req.CategoryIDs,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	if result == nil {
		h.internalServerError(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product": ToResponse(result)})
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateProduct(c.Request.Context(), id, &ports.UpdateProductDTO{
		Name: req.Name, Code: req.Code, SKU: req.SKU, Barcode: req.Barcode,
		Description: req.Description, IsActive: req.IsActive, CategoryIDs: req.CategoryIDs,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	if result == nil {
		h.internalServerError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": ToResponse(result)})
}

func (h *Handler) UpdateProductStatus(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_active is required"})
		return
	}

	if err := h.service.UpdateProductStatus(
		c.Request.Context(),
		id,
		*req.IsActive,
	); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "product status updated successfully",
	})
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "product deleted successfully"})
}

func (h *Handler) GetProductByID(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}
	result, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if result == nil {
		h.internalServerError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": ToResponse(result)})
}

func (h *Handler) ListProducts(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}

	results, total, err := h.service.ListProducts(c.Request.Context(), ports.ListProductsDTO{
		Page: query.Page, PageSize: query.PageSize, Search: query.Search,
		IsActive: query.IsActive, CategoryID: query.CategoryID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]ProductResponse, 0, len(results))
	for _, result := range results {
		if result == nil {
			h.internalServerError(c)
			return
		}
		items = append(items, ToResponse(result))
	}
	c.JSON(http.StatusOK, ListResponse{
		Items: items, Total: total, Page: query.Page, PageSize: query.PageSize,
	})
}

func (h *Handler) GetPublicProductByID(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}
	result, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if result == nil || !result.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrProductNotFound.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": ToPublicResponse(result)})
}

func (h *Handler) ListPublicProducts(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}

	active := true
	results, total, err := h.service.ListProducts(c.Request.Context(), ports.ListProductsDTO{
		Page: query.Page, PageSize: query.PageSize, Search: query.Search,
		IsActive: &active, CategoryID: query.CategoryID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]PublicProductResponse, 0, len(results))
	for _, result := range results {
		if result == nil {
			h.internalServerError(c)
			return
		}
		items = append(items, ToPublicResponse(result))
	}
	c.JSON(http.StatusOK, PublicListResponse{
		Items: items, Total: total, Page: query.Page, PageSize: query.PageSize,
	})
}

func parseProductID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return 0, false
	}
	return id, true
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": domain.ErrProductNotFound.Error()})
	case errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrProductNameTooLong),
		errors.Is(err, domain.ErrProductCodeRequired),
		errors.Is(err, domain.ErrProductCodeTooLong),
		errors.Is(err, domain.ErrProductSKUTooLong),
		errors.Is(err, domain.ErrProductBarcodeTooLong),
		errors.Is(err, domain.ErrProductDescriptionTooLong),
		errors.Is(err, domain.ErrProductCategoryInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrProductCodeAlreadyExists),
		errors.Is(err, domain.ErrProductSKUAlreadyExists),
		errors.Is(err, domain.ErrProductBarcodeAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		h.internalServerError(c)
	}
}

func (h *Handler) internalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
