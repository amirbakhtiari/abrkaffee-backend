package product_pricing_service

import (
	"context"
	"errors"

	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	pricingports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/pricing"
)

type service struct {
	productService ports.ProductService
	pricingService pricingports.PricingService
}

func NewProductPricingService(
	productService ports.ProductService,
	pricingService pricingports.PricingService,
) ports.ProductPricingService {
	return &service{
		productService: productService,
		pricingService: pricingService,
	}
}

func (s *service) ListProductsWithPrices(
	ctx context.Context,
	dto ports.ListProductsWithPricesDTO,
) ([]*ports.ProductWithPriceDTO, int64, error) {
	if dto.Page < 1 {
		dto.Page = 1
	}

	if dto.PageSize < 1 {
		dto.PageSize = 20
	}

	if dto.PageSize > 100 {
		dto.PageSize = 100
	}

	products, total, err := s.productService.ListProducts(
		ctx,
		ports.ListProductsDTO{
			Page:       dto.Page,
			PageSize:   dto.PageSize,
			Search:     dto.Search,
			IsActive:   dto.IsActive,
			CategoryID: dto.CategoryID,
		},
	)
	if err != nil {
		return nil, 0, err
	}

	if len(products) == 0 {
		return []*ports.ProductWithPriceDTO{}, total, nil
	}

	productIDs := make([]uint64, 0, len(products))

	for _, product := range products {
		if product == nil {
			return nil, 0, errors.New("product service returned a nil product")
		}

		productIDs = append(productIDs, product.ID)
	}

	prices, err := s.pricingService.GetActivePricesByProductIDs(
		ctx,
		productIDs,
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*ports.ProductWithPriceDTO, 0, len(products))

	for _, product := range products {
		item := &ports.ProductWithPriceDTO{
			Product: *product,
		}

		if activePrice, exists := prices[product.ID]; exists {
			item.Price = &ports.ProductPriceDTO{
				AnnouncementItemID: activePrice.AnnouncementItemID,
				AnnouncementID:     activePrice.AnnouncementID,
				Price:              activePrice.Price,
				AllowPriceChange:   activePrice.AllowPriceChange,
				MaxDiscountPercent: activePrice.MaxDiscountPercent,
				ApprovedAt:         activePrice.ApprovedAt,
			}
		}

		result = append(result, item)
	}

	return result, total, nil
}
