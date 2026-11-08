package pricing_service

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
	pricingport "github.com/amirbakhtiari/abrcoffee/internal/core/ports/pricing"
)

type service struct {
	repo pricingport.PricingRepository
}

func NewPricingService(repo pricingport.PricingRepository) pricingport.PricingService {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateAnnouncement(ctx context.Context, dto pricingport.CreateAnnouncementDTO) (*pricing.Announcement, error) {
	now := time.Now()

	domainItems := make([]pricing.Item, 0, len(dto.Items))
	for _, it := range dto.Items {
		domainItems = append(domainItems, pricing.NewItem(
			it.ProductID,
			it.Price,
			it.AllowPriceChange,
			it.MaxDiscountPercent,
		))
	}

	registeredAt := dto.RegisteredAt
	if registeredAt.IsZero() {
		registeredAt = now
	}

	announcement, err := pricing.NewAnnouncement(
		dto.Title,
		dto.Code,
		registeredAt,
		domainItems,
		now,
	)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, announcement)
}

func (s *service) GetAnnouncementByID(ctx context.Context, id uint64) (*pricing.Announcement, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ApproveAnnouncement(ctx context.Context, id uint64) error {
	announcement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := announcement.Approve(now); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, announcement)
}

func (s *service) RejectAnnouncement(ctx context.Context, id uint64) error {
	announcement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := announcement.Reject(now); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, announcement)
}

func (s *service) GetActiveProductPrice(ctx context.Context, productID uint64) (*pricingport.ActiveProductPrice, error) {
	return s.repo.GetActivePrice(ctx, productID)
}

func (s *service) GetActivePricesByProductIDs(ctx context.Context, productIDs []uint64) (map[uint64]pricingport.ActiveProductPrice, error) {
	return s.repo.GetActivePricesByProductIDs(ctx, productIDs)
}

func (s *service) ListAnnouncements(ctx context.Context, filter pricingport.ListAnnouncementsFilter) ([]*pricing.Announcement, int64, error) {
	return s.repo.List(ctx, filter)
}
