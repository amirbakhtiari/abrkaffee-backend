package product_pricing

import (
	"net/http"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ports.ProductPricingService
}

func NewHandler(service ports.ProductPricingService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) ListProductsWithPrices(c *gin.Context) {
	query, err := parseListQuery(c, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	results, total, err := h.service.ListProductsWithPrices(
		c.Request.Context(),
		query,
	)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve products with prices",
		})
		return
	}

	items := make([]ProductWithPriceResponse, 0, len(results))

	for _, result := range results {
		if result == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid product pricing data",
			})
			return
		}

		items = append(items, toProductWithPriceResponse(result))
	}

	c.JSON(http.StatusOK, ListProductsWithPricesResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
}

func (h *Handler) ListPublicProductsWithPrices(c *gin.Context) {
	query, err := parseListQuery(c, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	results, total, err := h.service.ListProductsWithPrices(
		c.Request.Context(),
		query,
	)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve products",
		})
		return
	}

	items := make([]PublicProductWithPriceResponse, 0, len(results))

	for _, result := range results {
		if result == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid product data",
			})
			return
		}

		items = append(items, toPublicProductWithPriceResponse(result))
	}

	c.JSON(http.StatusOK, PublicListProductsWithPricesResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
}
