package gateways

import (
	"context"

	categoryports "github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	productports "github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	orderports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/order"
	pricingports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/pricing"
)

type ProductCatalogGateway struct {
	productRepo  productports.ProductRepository
	categoryRepo productports.CategoryRepository
}

func NewProductCatalogGateway(
	productRepo productports.ProductRepository,
	categoryRepo categoryports.CategoryRepository,
) orderports.ProductCatalogPort {
	return &ProductCatalogGateway{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (g *ProductCatalogGateway) GetActiveProductsByIDs(ctx context.Context, productIDs []uint64) (map[uint64]orderports.ProductSnapshot, error) {
	result := make(map[uint64]orderports.ProductSnapshot, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	for _, id := range productIDs {
		p, err := g.productRepo.GetByID(ctx, id)
		if err != nil {
			// اگر محصول یافت نشد یا خطا داشت، آن را نادیده می‌گیریم تا در لیست نیاید
			// و سرویس سفارش نبودِ آن را تشخیص دهد
			continue
		}
		if p == nil {
			continue
		}

		var catID uint64
		var catTitle string

		catIDs := p.CategoryIDs()
		if len(catIDs) > 0 {
			catID = catIDs[0]
			if g.categoryRepo != nil {
				if cat, err := g.categoryRepo.GetByID(ctx, catID); err == nil && cat != nil {
					catTitle = cat.Name()
				}
			}
		}

		result[id] = orderports.ProductSnapshot{
			ID:            p.ID(),
			Name:          p.Name(),
			Code:          p.Code(),
			CategoryID:    catID,
			CategoryTitle: catTitle,
			IsActive:      p.IsActive(),
		}
	}

	return result, nil
}

type PricingCatalogGateway struct {
	pricingRepo pricingports.PricingRepository
}

func NewPricingCatalogGateway(repo pricingports.PricingRepository) orderports.PricingCatalogPort {
	return &PricingCatalogGateway{
		pricingRepo: repo,
	}
}

func (g *PricingCatalogGateway) GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]orderports.ActivePriceSnapshot, error) {
	result := make(map[uint64]orderports.ActivePriceSnapshot, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	prices, err := g.pricingRepo.GetActivePricesByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	for productID, p := range prices {
		result[productID] = orderports.ActivePriceSnapshot{
			AnnouncementItemID: 0,
			Price:              p.Price,
		}
	}

	return result, nil
}
