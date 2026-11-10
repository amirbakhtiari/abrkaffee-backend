package pg_pricing

import (
	"context"
	"errors"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
	pricingports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/pricing"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) pricingports.PricingRepository {
	return &repository{db: db}
}

func isPostgresErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}

func (r *repository) Create(ctx context.Context, a *pricing.Announcement) (*pricing.Announcement, error) {
	header := PriceAnnouncementEntity{
		Title:        a.Title(),
		Code:         a.Code(),
		RegisteredAt: a.RegisteredAt(),
		Status:       string(a.Status()),
		ApprovedAt:   a.ApprovedAt(),
		CreatedAt:    a.CreatedAt(),
		UpdatedAt:    a.UpdatedAt(),
	}

	for _, item := range a.Items() {
		header.Items = append(header.Items, PriceAnnouncementItemEntity{
			ProductID:          item.ProductID(),
			Price:              item.Price(),
			AllowPriceChange:   item.AllowPriceChange(),
			MaxDiscountPercent: item.MaxDiscountPercent(),
			CreatedAt:          a.CreatedAt(),
			UpdatedAt:          a.UpdatedAt(),
		})
	}

	err := r.db.WithContext(ctx).Create(&header).Error
	if err != nil {
		if isPostgresErrorCode(err, "23505") {
			return nil, pricing.ErrAnnouncementCodeDuplicate
		}
		if isPostgresErrorCode(err, "23503") {
			return nil, pricing.ErrProductNotFound
		}
		return nil, err
	}

	return toDomain(&header), nil
}

func (r *repository) GetByID(ctx context.Context, id uint64) (*pricing.Announcement, error) {
	var entity PriceAnnouncementEntity
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pricing.ErrAnnouncementNotFound
		}
		return nil, err
	}

	return toDomain(&entity), nil
}

func (r *repository) UpdateStatus(ctx context.Context, a *pricing.Announcement) error {
	res := r.db.WithContext(ctx).
		Model(&PriceAnnouncementEntity{}).
		Where("id = ? AND deleted_at IS NULL", a.ID()).
		Updates(map[string]interface{}{
			"status":      string(a.Status()),
			"approved_at": a.ApprovedAt(),
			"updated_at":  a.UpdatedAt(),
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return pricing.ErrAnnouncementNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, filter pricingports.ListAnnouncementsFilter) ([]*pricing.Announcement, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit

	query := r.db.WithContext(ctx).Model(&PriceAnnouncementEntity{}).Where("deleted_at IS NULL")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entities []PriceAnnouncementEntity
	err := query.
		Order("id DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&entities).Error

	if err != nil {
		return nil, 0, err
	}

	result := make([]*pricing.Announcement, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toDomain(&entity))
	}

	return result, total, nil
}

func (r *repository) GetActivePrice(ctx context.Context, productID uint64) (*pricingports.ActiveProductPrice, error) {
	type result struct {
		AnnouncementItemID uint64
		AnnouncementID     uint64
		Price              int64
		AllowPriceChange   bool
		MaxDiscountPercent float64
		ApprovedAt         time.Time
	}

	var res result
	err := r.db.WithContext(ctx).
		Table("price_announcement_items as pai").
		Select(`
    		pai.id AS announcement_item_id,
    		pa.id AS announcement_id,
    		pai.price,
    		pai.allow_price_change,
    		pai.max_discount_percent,
    		pa.approved_at
		`).
		Joins("JOIN price_announcements as pa ON pa.id = pai.announcement_id").
		Where("pai.product_id = ?", productID).
		Where("pa.status = ?", string(pricing.StatusApproved)).
		Where("pa.deleted_at IS NULL").
		Order("pa.approved_at DESC, pa.id DESC").
		Limit(1).
		Scan(&res).Error

	if err != nil {
		return nil, err
	}
	if res.Price == 0 {
		return nil, nil
	}

	return &pricingports.ActiveProductPrice{
		AnnouncementItemID: res.AnnouncementItemID,
		AnnouncementID:     res.AnnouncementID,
		Price:              res.Price,
		AllowPriceChange:   res.AllowPriceChange,
		MaxDiscountPercent: res.MaxDiscountPercent,
		ApprovedAt:         res.ApprovedAt,
	}, nil
}

func (r *repository) GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]pricingports.ActiveProductPrice, error) {
	if len(productIDs) == 0 {
		return make(map[uint64]pricingports.ActiveProductPrice), nil
	}

	type result struct {
		ProductID          uint64
		AnnouncementItemID uint64
		AnnouncementID     uint64
		Price              int64
		AllowPriceChange   bool
		MaxDiscountPercent float64
		ApprovedAt         time.Time
	}
	var results []result
	err := r.db.WithContext(ctx).
		Table("price_announcement_items as pai").
		Select(`
    		DISTINCT ON (pai.product_id)
    		pai.product_id,
    		pai.id AS announcement_item_id,
    		pa.id AS announcement_id,
    		pai.price,
    		pai.allow_price_change,
    		pai.max_discount_percent,
    		pa.approved_at
		`).
		Joins("JOIN price_announcements as pa ON pa.id = pai.announcement_id").
		Where("pai.product_id IN ?", productIDs).
		Where("pa.status = ?", string(pricing.StatusApproved)).
		Where("pa.deleted_at IS NULL").
		Order("pai.product_id, pa.approved_at DESC, pa.id DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	priceMap := make(map[uint64]pricingports.ActiveProductPrice, len(results))

	for _, row := range results {
		priceMap[row.ProductID] = pricingports.ActiveProductPrice{
			AnnouncementItemID: row.AnnouncementItemID,
			AnnouncementID:     row.AnnouncementID,
			Price:              row.Price,
			AllowPriceChange:   row.AllowPriceChange,
			MaxDiscountPercent: row.MaxDiscountPercent,
			ApprovedAt:         row.ApprovedAt,
		}
	}

	return priceMap, nil
}

func toDomain(e *PriceAnnouncementEntity) *pricing.Announcement {
	items := make([]pricing.Item, 0, len(e.Items))
	for _, it := range e.Items {
		items = append(items, pricing.ReconstructItem(
			it.ID,
			it.AnnouncementID,
			it.ProductID,
			it.Product.Name,
			it.Product.Code,
			it.Price,
			it.AllowPriceChange,
			it.MaxDiscountPercent,
		))
	}

	return pricing.ReconstructAnnouncement(
		e.ID,
		e.Title,
		e.Code,
		e.RegisteredAt,
		pricing.Status(e.Status),
		e.ApprovedAt,
		e.CreatedAt,
		e.UpdatedAt,
		items,
	)
}
