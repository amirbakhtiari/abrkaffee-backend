package pg_order

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/order"
	orderports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/order"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) orderports.OrderRepository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uint64) (*domain.Order, error) {
	var ent OrderEntity
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&ent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	return toDomainOrder(&ent), nil
}

func (r *repository) GetByIdempotencyKey(ctx context.Context, key uuid.UUID) (*domain.Order, error) {
	var ent OrderEntity
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("idempotency_key = ? AND deleted_at IS NULL", key).
		First(&ent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toDomainOrder(&ent), nil
}

func (r *repository) CreateAtomic(ctx context.Context, o *domain.Order) (*domain.Order, error) {
	var createdOrder *domain.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// بررسی مجدد کلید Idempotency درون تراکنش
		if o.IdempotencyKey() != nil {
			var existing OrderEntity
			res := tx.Where("idempotency_key = ? AND deleted_at IS NULL", *o.IdempotencyKey()).
				Preload("Items").
				Limit(1).
				Find(&existing)
			if res.Error == nil && res.RowsAffected > 0 {
				createdOrder = toDomainOrder(&existing)
				return nil
			}
		}

		// دریافت عدد نوبت روزانه به صورت Atomic
		var nextSeq int64
		seqSQL := `
			INSERT INTO order_daily_sequences (order_date, last_number)
			VALUES (?, 1)
			ON CONFLICT (order_date)
			DO UPDATE SET last_number = order_daily_sequences.last_number + 1
			RETURNING last_number;
		`
		dateStr := o.OrderDate().Format("2006-01-02")
		if err := tx.Raw(seqSQL, dateStr).Scan(&nextSeq).Error; err != nil {
			return fmt.Errorf("failed to increment daily sequence: %w", err)
		}

		o.SetDailyNumber(uint64(nextSeq))

		// درج سربرگ سفارش
		orderEnt := OrderEntity{
			IdempotencyKey: o.IdempotencyKey(),
			OrderDate:      o.OrderDate(),
			DailyNumber:    o.DailyNumber(),
			TableID:        o.TableID(),
			CustomerPhone:  o.CustomerPhone(),
			CustomerName:   o.CustomerName(),
			CustomerNote:   o.CustomerNote(),
			Status:         string(o.Status()),
			PaymentStatus:  string(o.PaymentStatus()),
			TotalAmount:    o.TotalAmount(),
			Version:        o.Version(),
			RegisteredAt:   o.RegisteredAt(),
			CreatedAt:      o.CreatedAt(),
			UpdatedAt:      o.UpdatedAt(),
		}

		if err := tx.Create(&orderEnt).Error; err != nil {
			return err
		}

		// درج اقلام سفارش
		itemEntities := make([]OrderItemEntity, 0, len(o.Items()))
		for _, item := range o.Items() {
			itemEntities = append(itemEntities, OrderItemEntity{
				OrderID:                 orderEnt.ID,
				ProductID:               item.ProductID(),
				CategoryID:              item.CategoryID(),
				PriceAnnouncementItemID: item.PriceAnnouncementItemID(),
				ProductName:             item.ProductName(),
				ProductCode:             item.ProductCode(),
				CategoryTitle:           item.CategoryTitle(),
				UnitPrice:               item.UnitPrice(),
				Quantity:                item.Quantity(),
				TotalPrice:              item.TotalPrice(),
				Notes:                   item.Notes(),
				CreatedAt:               orderEnt.CreatedAt,
			})
		}

		if err := tx.Create(&itemEntities).Error; err != nil {
			return err
		}

		// درج وضعیت اولیه در جدول تاریخچه
		hist := OrderStatusHistoryEntity{
			OrderID:    orderEnt.ID,
			FromStatus: nil,
			ToStatus:   string(domain.StatusPending),
			CreatedAt:  orderEnt.CreatedAt,
		}
		if err := tx.Create(&hist).Error; err != nil {
			return err
		}

		orderEnt.Items = itemEntities
		createdOrder = toDomainOrder(&orderEnt)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

func (r *repository) UpdateStatus(ctx context.Context, o *domain.Order, fromStatus domain.Status, changedBy *uint64, note *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// آپدیت با Optimistic Lock
		fromStr := string(fromStatus)
		res := tx.Model(&OrderEntity{}).
			Where("id = ? AND version = ? AND deleted_at IS NULL", o.ID(), o.Version()).
			Updates(map[string]interface{}{
				"status":     string(o.Status()),
				"version":    o.Version() + 1,
				"updated_at": o.UpdatedAt(),
			})

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrOptimisticLockConflict
		}

		// ثبت در تاریخچه
		hist := OrderStatusHistoryEntity{
			OrderID:    o.ID(),
			FromStatus: &fromStr,
			ToStatus:   string(o.Status()),
			ChangedBy:  changedBy,
			Note:       note,
			CreatedAt:  time.Now(),
		}
		return tx.Create(&hist).Error
	})
}

func (r *repository) AddPayment(ctx context.Context, o *domain.Order, p *domain.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pEnt := OrderPaymentEntity{
			OrderID:              o.ID(),
			Amount:               p.Amount(),
			PaymentMethod:        string(p.PaymentMethod()),
			Status:               string(p.Status()),
			TransactionReference: p.TransactionReference(),
			PaidAt:               p.PaidAt(),
			CreatedAt:            time.Now(),
		}
		if err := tx.Create(&pEnt).Error; err != nil {
			return err
		}

		res := tx.Model(&OrderEntity{}).
			Where("id = ? AND deleted_at IS NULL", o.ID()).
			Updates(map[string]interface{}{
				"payment_status": string(o.PaymentStatus()),
				"settled_at":     o.SettledAt(),
				"updated_at":     o.UpdatedAt(),
			})

		return res.Error
	})
}

func (r *repository) List(ctx context.Context, filter orderports.ListOrdersFilter) ([]*domain.Order, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit

	query := r.db.WithContext(ctx).Model(&OrderEntity{}).Where("deleted_at IS NULL")

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}
	if filter.PaymentStatus != nil {
		query = query.Where("payment_status = ?", string(*filter.PaymentStatus))
	}
	if filter.Date != nil {
		query = query.Where("order_date = ?", filter.Date.Format("2006-01-02"))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entities []OrderEntity
	err := query.
		Preload("Items").
		Order("id DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&entities).Error

	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Order, 0, len(entities))
	for _, ent := range entities {
		result = append(result, toDomainOrder(&ent))
	}

	return result, total, nil
}

func toDomainOrder(e *OrderEntity) *domain.Order {
	items := make([]domain.Item, 0, len(e.Items))
	for _, it := range e.Items {
		items = append(items, domain.ReconstructItem(
			it.ID,
			it.OrderID,
			it.ProductID,
			it.CategoryID,
			it.PriceAnnouncementItemID,
			it.ProductName,
			it.ProductCode,
			it.CategoryTitle,
			it.UnitPrice,
			it.Quantity,
			it.TotalPrice,
			it.Notes,
		))
	}

	return domain.ReconstructOrder(
		e.ID,
		e.IdempotencyKey,
		e.OrderDate,
		e.DailyNumber,
		e.TableID,
		e.CustomerPhone,
		e.CustomerName,
		e.CustomerNote,
		domain.Status(e.Status),
		domain.PaymentStatus(e.PaymentStatus),
		e.TotalAmount,
		e.Version,
		e.RegisteredAt,
		e.SettledAt,
		e.CreatedAt,
		e.UpdatedAt,
		items,
	)
}
