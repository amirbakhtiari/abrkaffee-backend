package order_service

import (
	"context"
	"fmt"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/order"
	orderports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/order"
)

type service struct {
	orderRepo   orderports.OrderRepository
	productPort orderports.ProductCatalogPort
	pricingPort orderports.PricingCatalogPort
}

func NewOrderService(
	orderRepo orderports.OrderRepository,
	productPort orderports.ProductCatalogPort,
	pricingPort orderports.PricingCatalogPort,
) orderports.OrderService {
	return &service{
		orderRepo:   orderRepo,
		productPort: productPort,
		pricingPort: pricingPort,
	}
}

func (s *service) CreateOrder(ctx context.Context, cmd orderports.CreateOrderCommand) (*domain.Order, error) {
	// ۱. گارد Idempotency
	if cmd.IdempotencyKey != nil {
		existing, err := s.orderRepo.GetByIdempotencyKey(ctx, *cmd.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	if len(cmd.Items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}

	// استخراج شناسه‌های محصولات
	productIDs := make([]uint64, 0, len(cmd.Items))
	for _, it := range cmd.Items {
		productIDs = append(productIDs, it.ProductID)
	}

	// ۲. استعلام مشخصات و وضعیت فعال بودن کالاها از Product Catalog
	products, err := s.productPort.GetActiveProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// ۳. استعلام آخرین قیمت مصوب از Pricing Port
	prices, err := s.pricingPort.GetActivePricesByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active prices: %w", err)
	}

	// ۴. اعتبارسنجی و ساخت Snapshot اقلام
	domainItems := make([]domain.Item, 0, len(cmd.Items))
	for _, itemParam := range cmd.Items {
		prod, exists := products[itemParam.ProductID]
		if !exists || !prod.IsActive {
			return nil, domain.ErrProductNotActive
		}

		priceInfo, priceExists := prices[itemParam.ProductID]
		if !priceExists || priceInfo.Price <= 0 {
			return nil, domain.ErrActivePriceNotFound
		}

		announcementItemID := priceInfo.AnnouncementItemID
		dItem, err := domain.NewItem(
			prod.ID,
			prod.CategoryID,
			&announcementItemID,
			prod.Name,
			prod.Code,
			prod.CategoryTitle,
			priceInfo.Price,
			itemParam.Quantity,
			itemParam.Notes,
		)
		if err != nil {
			return nil, err
		}
		domainItems = append(domainItems, dItem)
	}

	// ۵. ساخت مدل دامین
	orderEntity, err := domain.NewOrder(
		cmd.IdempotencyKey,
		cmd.OrderDate,
		cmd.TableID,
		cmd.CustomerPhone,
		cmd.CustomerName,
		cmd.CustomerNote,
		domainItems,
	)
	if err != nil {
		return nil, err
	}

	// ۶. اجرای تراکنش اتمیک
	return s.orderRepo.CreateAtomic(ctx, orderEntity)
}

func (s *service) GetOrderByID(ctx context.Context, id uint64) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *service) UpdateOrderStatus(ctx context.Context, orderID uint64, targetStatus domain.Status, changedBy *uint64, note *string) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	currentStatus := o.Status()
	switch targetStatus {
	case domain.StatusPreparing:
		if err := o.MarkPreparing(); err != nil {
			return err
		}
	case domain.StatusServed:
		if err := o.MarkServed(); err != nil {
			return err
		}
	case domain.StatusCompleted:
		if err := o.Complete(); err != nil {
			return err
		}
	case domain.StatusCanceled:
		if err := o.Cancel(); err != nil {
			return err
		}
	default:
		return domain.ErrInvalidStatusTransition
	}

	return s.orderRepo.UpdateStatus(ctx, o, currentStatus, changedBy, note)
}

func (s *service) RecordPayment(ctx context.Context, orderID uint64, amount int64, method domain.PaymentMethod, ref *string) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if o.PaymentStatus() == domain.PaymentStatusPaid {
		return domain.ErrOrderAlreadyCompleted
	}
	if amount > o.TotalAmount() {
		return domain.ErrPaymentExceedsRemaining
	}

	payment, err := domain.NewPayment(o.ID(), amount, method, ref)
	if err != nil {
		return err
	}

	// بررسی وضعیت پرداخت
	if amount == o.TotalAmount() {
		o.UpdatePaymentStatus(domain.PaymentStatusPaid)
	} else {
		o.UpdatePaymentStatus(domain.PaymentStatusPartiallyPaid)
	}

	return s.orderRepo.AddPayment(ctx, o, payment)
}

func (s *service) ListOrders(ctx context.Context, filter orderports.ListOrdersFilter) ([]*domain.Order, int64, error) {
	return s.orderRepo.List(ctx, filter)
}
