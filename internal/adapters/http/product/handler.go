package product

import (
	"encoding/json"
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 6<<20)

	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		if isRequestTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body is too large",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid multipart request",
		})
		return
	}

	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "image is required",
		})
		return
	}
	defer file.Close()

	if header.Size > 5<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "image must not exceed 5 MB",
		})
		return
	}

	categoryIDs, err := parseCategoryIDs(c.PostForm("category_ids"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category_ids",
		})
		return
	}

	isActive := true
	if raw := c.PostForm("is_active"); raw != "" {
		parsed, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid is_active",
			})
			return
		}
		isActive = parsed
	}

	result, err := h.service.CreateProduct(
		c.Request.Context(),
		&ports.CreateProductDTO{
			Name:        c.PostForm("name"),
			Code:        c.PostForm("code"),
			SKU:         c.PostForm("sku"),
			Barcode:     c.PostForm("barcode"),
			Description: c.PostForm("description"),
			IsActive:    isActive,
			CategoryIDs: categoryIDs,
			Image:       file,
			ImageType:   header.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	if result == nil {
		h.internalServerError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"product": ToResponse(result),
	})
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
	case errors.Is(err, domain.ErrProductImageRequired),
		errors.Is(err, ports.ErrUnsupportedFileType):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ports.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	default:
		h.internalServerError(c)
	}
}

func (h *Handler) internalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func parseCategoryIDs(raw string) ([]uint64, error) {
	if raw == "" {
		return nil, nil
	}

	var ids []uint64
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func isRequestTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

func (h *Handler) UpdateProductImage(c *gin.Context) {
	id, ok := parseProductID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 6<<20)

	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		if isRequestTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body is too large",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid multipart request",
		})
		return
	}

	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "image is required",
		})
		return
	}
	defer file.Close()

	if header.Size > 5<<20 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "image must not exceed 5 MB",
		})
		return
	}

	if err := h.service.UpdateProductImage(
		c.Request.Context(),
		id,
		file,
		header.Header.Get("Content-Type"),
	); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "product image updated successfully",
	})
}
