package product_pricing

import (
	"fmt"
	"strconv"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/gin-gonic/gin"
)

func parseListQuery(c *gin.Context, public bool) (ports.ListProductsWithPricesDTO, error) {
	query := ports.ListProductsWithPricesDTO{
		Page:     1,
		PageSize: 20,
		Search:   c.Query("search"),
	}

	if raw := c.Query("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 1 {
			return query, fmt.Errorf("invalid page")
		}
		query.Page = page
	}

	if raw := c.Query("page_size"); raw != "" {
		pageSize, err := strconv.Atoi(raw)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return query, fmt.Errorf("invalid page_size")
		}
		query.PageSize = pageSize
	}

	if raw := c.Query("category_id"); raw != "" {
		categoryID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || categoryID == 0 {
			return query, fmt.Errorf("invalid category_id")
		}
		query.CategoryID = categoryID
	}

	if public {
		active := true
		query.IsActive = &active
	} else if raw := c.Query("is_active"); raw != "" {
		isActive, err := strconv.ParseBool(raw)
		if err != nil {
			return query, fmt.Errorf("invalid is_active")
		}
		query.IsActive = &isActive
	}

	return query, nil
}
