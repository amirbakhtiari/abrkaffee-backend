package order

import (
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/order"
)

type OrderItemResponse struct {
	ID            uint64  `json:"id"`
	ProductID     uint64  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	ProductCode   string  `json:"product_code"`
	CategoryID    uint64  `json:"category_id"`
	CategoryTitle string  `json:"category_title"`
	UnitPrice     int64   `json:"unit_price"`
	Quantity      int     `json:"quantity"`
	TotalPrice    int64   `json:"total_price"`
	Notes         *string `json:"notes,omitempty"`
}

type OrderResponse struct {
	ID            uint64              `json:"id"`
	DailyNumber   uint64              `json:"daily_number"`
	OrderDate     string              `json:"order_date"`
	TableID       uint64              `json:"table_id"`
	CustomerPhone *string             `json:"customer_phone,omitempty"`
	CustomerName  *string             `json:"customer_name,omitempty"`
	CustomerNote  *string             `json:"customer_note,omitempty"`
	Status        string              `json:"status"`
	PaymentStatus string              `json:"payment_status"`
	TotalAmount   int64               `json:"total_amount"`
	RegisteredAt  time.Time           `json:"registered_at"`
	SettledAt     *time.Time          `json:"settled_at,omitempty"`
	Items         []OrderItemResponse `json:"items"`
}

func ToOrderResponse(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items()))
	for _, it := range o.Items() {
		items = append(items, OrderItemResponse{
			ID:            it.ID(),
			ProductID:     it.ProductID(),
			ProductName:   it.ProductName(),
			ProductCode:   it.ProductCode(),
			CategoryID:    it.CategoryID(),
			CategoryTitle: it.CategoryTitle(),
			UnitPrice:     it.UnitPrice(),
			Quantity:      it.Quantity(),
			TotalPrice:    it.TotalPrice(),
			Notes:         it.Notes(),
		})
	}

	return OrderResponse{
		ID:            o.ID(),
		DailyNumber:   o.DailyNumber(),
		OrderDate:     o.OrderDate().Format("2006-01-02"),
		TableID:       o.TableID(),
		CustomerPhone: o.CustomerPhone(),
		CustomerName:  o.CustomerName(),
		CustomerNote:  o.CustomerNote(),
		Status:        string(o.Status()),
		PaymentStatus: string(o.PaymentStatus()),
		TotalAmount:   o.TotalAmount(),
		RegisteredAt:  o.RegisteredAt(),
		SettledAt:     o.SettledAt(),
		Items:         items,
	}
}
