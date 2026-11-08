package order

import (
	"github.com/google/uuid"
)

type CreateOrderItemRequest struct {
	ProductID uint64  `json:"product_id" binding:"required,gt=0"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	Notes     *string `json:"notes,omitempty"`
}

type CreateOrderRequest struct {
	IdempotencyKey *uuid.UUID               `json:"idempotency_key,omitempty"`
	TableID        uint64                   `json:"table_id" binding:"required,gt=0"`
	CustomerPhone  *string                  `json:"customer_phone,omitempty"`
	CustomerName   *string                  `json:"customer_name,omitempty"`
	CustomerNote   *string                  `json:"customer_note,omitempty"`
	Items          []CreateOrderItemRequest `json:"items" binding:"required,dive"`
}

type UpdateOrderStatusRequest struct {
	Status string  `json:"status" binding:"required"`
	Note   *string `json:"note,omitempty"`
}

type RecordPaymentRequest struct {
	Amount        int64   `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=CASH POS ONLINE"`
	Reference     *string `json:"reference,omitempty"`
}
