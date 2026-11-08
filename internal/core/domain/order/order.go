package order

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPreparing Status = "PREPARING"
	StatusServed    Status = "SERVED"
	StatusCompleted Status = "COMPLETED"
	StatusCanceled  Status = "CANCELED"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid        PaymentStatus = "UNPAID"
	PaymentStatusPartiallyPaid PaymentStatus = "PARTIALLY_PAID"
	PaymentStatusPaid          PaymentStatus = "PAID"
	PaymentStatusRefunded      PaymentStatus = "REFUNDED"
)

type Order struct {
	id             uint64
	idempotencyKey *uuid.UUID
	orderDate      time.Time
	dailyNumber    uint64
	tableID        uint64
	customerPhone  *string
	customerName   *string
	customerNote   *string
	status         Status
	paymentStatus  PaymentStatus
	totalAmount    int64
	version        int
	registeredAt   time.Time
	settledAt      *time.Time
	createdAt      time.Time
	updatedAt      time.Time
	items          []Item
}

func NewOrder(
	idempotencyKey *uuid.UUID,
	orderDate time.Time,
	tableID uint64,
	customerPhone *string,
	customerName *string,
	customerNote *string,
	items []Item,
) (*Order, error) {
	if tableID == 0 {
		return nil, ErrInvalidTableID
	}
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	var total int64
	for _, it := range items {
		total += it.TotalPrice()
	}

	now := time.Now()
	return &Order{
		idempotencyKey: idempotencyKey,
		orderDate:      orderDate,
		tableID:        tableID,
		customerPhone:  customerPhone,
		customerName:   customerName,
		customerNote:   customerNote,
		status:         StatusPending,
		paymentStatus:  PaymentStatusUnpaid,
		totalAmount:    total,
		version:        1,
		registeredAt:   now,
		createdAt:      now,
		updatedAt:      now,
		items:          items,
	}, nil
}

func ReconstructOrder(
	id uint64,
	idempotencyKey *uuid.UUID,
	orderDate time.Time,
	dailyNumber uint64,
	tableID uint64,
	customerPhone *string,
	customerName *string,
	customerNote *string,
	status Status,
	paymentStatus PaymentStatus,
	totalAmount int64,
	version int,
	registeredAt time.Time,
	settledAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	items []Item,
) *Order {
	return &Order{
		id:             id,
		idempotencyKey: idempotencyKey,
		orderDate:      orderDate,
		dailyNumber:    dailyNumber,
		tableID:        tableID,
		customerPhone:  customerPhone,
		customerName:   customerName,
		customerNote:   customerNote,
		status:         status,
		paymentStatus:  paymentStatus,
		totalAmount:    totalAmount,
		version:        version,
		registeredAt:   registeredAt,
		settledAt:      settledAt,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		items:          items,
	}
}

// Getters
func (o *Order) ID() uint64                   { return o.id }
func (o *Order) IdempotencyKey() *uuid.UUID   { return o.idempotencyKey }
func (o *Order) OrderDate() time.Time         { return o.orderDate }
func (o *Order) DailyNumber() uint64          { return o.dailyNumber }
func (o *Order) TableID() uint64              { return o.tableID }
func (o *Order) CustomerPhone() *string       { return o.customerPhone }
func (o *Order) CustomerName() *string        { return o.customerName }
func (o *Order) CustomerNote() *string        { return o.customerNote }
func (o *Order) Status() Status               { return o.status }
func (o *Order) PaymentStatus() PaymentStatus { return o.paymentStatus }
func (o *Order) TotalAmount() int64           { return o.totalAmount }
func (o *Order) Version() int                 { return o.version }
func (o *Order) RegisteredAt() time.Time      { return o.registeredAt }
func (o *Order) SettledAt() *time.Time        { return o.settledAt }
func (o *Order) CreatedAt() time.Time         { return o.createdAt }
func (o *Order) UpdatedAt() time.Time         { return o.updatedAt }
func (o *Order) Items() []Item                { return o.items }

func (o *Order) SetDailyNumber(n uint64) {
	o.dailyNumber = n
}

// State Machine transitions
func (o *Order) MarkPreparing() error {
	if o.status != StatusPending {
		return ErrInvalidStatusTransition
	}
	o.status = StatusPreparing
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) MarkServed() error {
	if o.status != StatusPreparing {
		return ErrInvalidStatusTransition
	}
	o.status = StatusServed
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) Complete() error {
	if o.status != StatusServed || o.paymentStatus != PaymentStatusPaid {
		return ErrCannotCompleteUnpaidOrder
	}
	o.status = StatusCompleted
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) Cancel() error {
	if o.status == StatusCompleted || o.status == StatusCanceled {
		return ErrOrderAlreadyCompleted
	}
	o.status = StatusCanceled
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) UpdatePaymentStatus(newStatus PaymentStatus) {
	o.paymentStatus = newStatus
	if newStatus == PaymentStatusPaid {
		now := time.Now()
		o.settledAt = &now
	}
	o.updatedAt = time.Now()
}
