package pricing

import (
	"strings"
	"time"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

type Item struct {
	id                 uint64
	announcementID     uint64
	productID          uint64
	productName        string
	productCode        string
	price              int64
	allowPriceChange   bool
	maxDiscountPercent float64
}

func NewItem(productID uint64, price int64, allowPriceChange bool, maxDiscountPercent float64) Item {
	return Item{
		productID:          productID,
		price:              price,
		allowPriceChange:   allowPriceChange,
		maxDiscountPercent: maxDiscountPercent,
	}
}

func ReconstructItem(
	id, announcementID, productID uint64,
	productName, productCode string,
	price int64,
	allowPriceChange bool,
	maxDiscountPercent float64,
) Item {
	return Item{
		id:                 id,
		announcementID:     announcementID,
		productID:          productID,
		productName:        productName,
		productCode:        productCode,
		price:              price,
		allowPriceChange:   allowPriceChange,
		maxDiscountPercent: maxDiscountPercent,
	}
}

func (i Item) ID() uint64                  { return i.id }
func (i Item) AnnouncementID() uint64      { return i.announcementID }
func (i Item) ProductID() uint64           { return i.productID }
func (i Item) Price() int64                { return i.price }
func (i Item) AllowPriceChange() bool      { return i.allowPriceChange }
func (i Item) MaxDiscountPercent() float64 { return i.maxDiscountPercent }
func (i Item) ProductName() string {
	return i.productName
}

func (i Item) ProductCode() string {
	return i.productCode
}

type Announcement struct {
	id           uint64
	title        string
	code         string
	registeredAt time.Time
	status       Status
	approvedAt   *time.Time
	createdAt    time.Time
	updatedAt    time.Time
	items        []Item
}

func NewAnnouncement(
	title, code string,
	registeredAt time.Time,
	rawItems []Item,
	now time.Time,
) (*Announcement, error) {
	title = strings.TrimSpace(title)
	code = strings.TrimSpace(code)

	if title == "" {
		return nil, ErrAnnouncementTitleRequired
	}
	if code == "" {
		return nil, ErrAnnouncementCodeRequired
	}
	if len(rawItems) == 0 {
		return nil, ErrNoItemsProvided
	}

	seenProducts := make(map[uint64]struct{}, len(rawItems))
	validItems := make([]Item, 0, len(rawItems))

	for _, it := range rawItems {
		if it.productID == 0 {
			continue
		}
		if _, exists := seenProducts[it.productID]; exists {
			return nil, ErrDuplicateProductInItems
		}
		if it.price <= 0 {
			return nil, ErrInvalidItemPrice
		}
		if it.maxDiscountPercent < 0 || it.maxDiscountPercent > 100 {
			return nil, ErrInvalidDiscountPercent
		}

		seenProducts[it.productID] = struct{}{}
		validItems = append(validItems, it)
	}

	if len(validItems) == 0 {
		return nil, ErrNoItemsProvided
	}

	return &Announcement{
		title:        title,
		code:         code,
		registeredAt: registeredAt,
		status:       StatusPending,
		createdAt:    now,
		updatedAt:    now,
		items:        validItems,
	}, nil
}

func ReconstructAnnouncement(
	id uint64,
	title, code string,
	registeredAt time.Time,
	status Status,
	approvedAt *time.Time,
	createdAt, updatedAt time.Time,
	items []Item,
) *Announcement {
	return &Announcement{
		id:           id,
		title:        title,
		code:         code,
		registeredAt: registeredAt,
		status:       status,
		approvedAt:   approvedAt,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		items:        items,
	}
}

func (a *Announcement) ID() uint64              { return a.id }
func (a *Announcement) Title() string           { return a.title }
func (a *Announcement) Code() string            { return a.code }
func (a *Announcement) RegisteredAt() time.Time { return a.registeredAt }
func (a *Announcement) Status() Status          { return a.status }
func (a *Announcement) ApprovedAt() *time.Time  { return a.approvedAt }
func (a *Announcement) CreatedAt() time.Time    { return a.createdAt }
func (a *Announcement) UpdatedAt() time.Time    { return a.updatedAt }
func (a *Announcement) Items() []Item           { return a.items }

func (a *Announcement) Approve(now time.Time) error {
	if a.status != StatusPending {
		return ErrAnnouncementAlreadyHandled
	}
	a.status = StatusApproved
	a.approvedAt = &now
	a.updatedAt = now
	return nil
}

func (a *Announcement) Reject(now time.Time) error {
	if a.status != StatusPending {
		return ErrAnnouncementAlreadyHandled
	}
	a.status = StatusRejected
	a.updatedAt = now
	return nil
}
