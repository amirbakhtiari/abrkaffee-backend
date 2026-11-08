package pg_order

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderDailySequenceEntity struct {
	OrderDate  time.Time `gorm:"primaryKey;type:date"`
	LastNumber int64     `gorm:"not null;default:0"`
}

func (OrderDailySequenceEntity) TableName() string {
	return "order_daily_sequences"
}

type OrderEntity struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"`
	IdempotencyKey *uuid.UUID     `gorm:"type:uuid;index"`
	OrderDate      time.Time      `gorm:"type:date;not null;index"`
	DailyNumber    uint64         `gorm:"not null"`
	TableID        uint64         `gorm:"not null;index"`
	CustomerPhone  *string        `gorm:"type:varchar(20)"`
	CustomerName   *string        `gorm:"type:varchar(100)"`
	CustomerNote   *string        `gorm:"type:text"`
	Status         string         `gorm:"type:varchar(20);not null;default:'PENDING'"`
	PaymentStatus  string         `gorm:"type:varchar(20);not null;default:'UNPAID'"`
	TotalAmount    int64          `gorm:"not null;default:0"`
	Version        int            `gorm:"not null;default:1"`
	RegisteredAt   time.Time      `gorm:"not null"`
	SettledAt      *time.Time     `gorm:""`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Items    []OrderItemEntity    `gorm:"foreignKey:OrderID"`
	Payments []OrderPaymentEntity `gorm:"foreignKey:OrderID"`
}

func (OrderEntity) TableName() string {
	return "orders"
}

type OrderItemEntity struct {
	ID                      uint64    `gorm:"primaryKey;autoIncrement"`
	OrderID                 uint64    `gorm:"not null;index"`
	ProductID               uint64    `gorm:"not null"`
	CategoryID              uint64    `gorm:"not null"`
	PriceAnnouncementItemID *uint64   `gorm:""`
	ProductName             string    `gorm:"type:varchar(255);not null"`
	ProductCode             string    `gorm:"type:varchar(64);not null"`
	CategoryTitle           string    `gorm:"type:varchar(100);not null"`
	UnitPrice               int64     `gorm:"not null"`
	Quantity                int       `gorm:"not null"`
	TotalPrice              int64     `gorm:"not null"`
	Notes                   *string   `gorm:"type:varchar(255)"`
	CreatedAt               time.Time `gorm:"not null"`
}

func (OrderItemEntity) TableName() string {
	return "order_items"
}

type OrderPaymentEntity struct {
	ID                   uint64     `gorm:"primaryKey;autoIncrement"`
	OrderID              uint64     `gorm:"not null;index"`
	Amount               int64      `gorm:"not null"`
	PaymentMethod        string     `gorm:"type:varchar(20);not null"`
	Status               string     `gorm:"type:varchar(20);not null"`
	TransactionReference *string    `gorm:"type:varchar(100)"`
	PaidAt               *time.Time `gorm:""`
	CreatedAt            time.Time  `gorm:"not null"`
}

func (OrderPaymentEntity) TableName() string {
	return "order_payments"
}

type OrderStatusHistoryEntity struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	OrderID    uint64    `gorm:"not null;index"`
	FromStatus *string   `gorm:"type:varchar(20)"`
	ToStatus   string    `gorm:"type:varchar(20);not null"`
	ChangedBy  *uint64   `gorm:""`
	Note       *string   `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"not null"`
}

func (OrderStatusHistoryEntity) TableName() string {
	return "order_status_histories"
}
