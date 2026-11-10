package product

import "context"

type PriceLookupPort interface {
	GetLatestPricesForProducts(ctx context.Context, productIDs []uint64) (map[uint64]int64, error)
}
