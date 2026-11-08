package order

import (
	"context"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/order"
	"github.com/google/uuid"
)

type ProductSnapshot struct {
	ID            uint64
	Name          string
	Code          string
	CategoryID    uint64
	CategoryTitle string
	IsActive      bool
}

type ActivePriceSnapshot struct {
	AnnouncementItemID uint64
	Price              int64
}

type ProductCatalogPort interface {
	GetActiveProductsByIDs(ctx context.Context, productIDs []uint64) (map[uint64]ProductSnapshot, error)
}

type PricingCatalogPort interface {
	GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]ActivePriceSnapshot, error)
}

type ListOrdersFilter struct {
	Page          int
	Limit         int
	TableID       *uint64
	Status        *domain.Status
	PaymentStatus *domain.PaymentStatus
	Date          *time.Time
}

// پورت ریپازیتوری
type OrderRepository interface {
	GetByID(ctx context.Context, id uint64) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (*domain.Order, error)
	CreateAtomic(ctx context.Context, o *domain.Order) (*domain.Order, error)
	UpdateStatus(ctx context.Context, o *domain.Order, fromStatus domain.Status, changedBy *uint64, note *string) error
	AddPayment(ctx context.Context, o *domain.Order, p *domain.Payment) error
	List(ctx context.Context, filter ListOrdersFilter) ([]*domain.Order, int64, error)
}

// پورت سرویس بیزینس
type CreateOrderItemParam struct {
	ProductID uint64
	Quantity  int
	Notes     *string
}

type CreateOrderCommand struct {
	IdempotencyKey *uuid.UUID
	OrderDate      time.Time
	TableID        uint64
	CustomerPhone  *string
	CustomerName   *string
	CustomerNote   *string
	Items          []CreateOrderItemParam
}

type OrderService interface {
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error)
	GetOrderByID(ctx context.Context, id uint64) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint64, targetStatus domain.Status, changedBy *uint64, note *string) error
	RecordPayment(ctx context.Context, orderID uint64, amount int64, method domain.PaymentMethod, ref *string) error
	ListOrders(ctx context.Context, filter ListOrdersFilter) ([]*domain.Order, int64, error)
}
