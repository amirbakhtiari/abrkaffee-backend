package order

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/order"
	orderports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/order"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service orderports.OrderService
}

func NewHandler(svc orderports.OrderService) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]orderports.CreateOrderItemParam, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, orderports.CreateOrderItemParam{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Notes:     it.Notes,
		})
	}

	// تنظیم مبنای تاریخ کاری (OrderDate) بر اساس زمان حال
	now := time.Now()
	orderDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	cmd := orderports.CreateOrderCommand{
		IdempotencyKey: req.IdempotencyKey,
		OrderDate:      orderDate,
		TableID:        req.TableID,
		CustomerPhone:  req.CustomerPhone,
		CustomerName:   req.CustomerName,
		CustomerNote:   req.CustomerNote,
		Items:          items,
	}

	order, err := h.service.CreateOrder(c.Request.Context(), cmd)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ToOrderResponse(order))
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToOrderResponse(order))
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateOrderStatus(c.Request.Context(), id, domain.Status(req.Status), nil, req.Note)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}

func (h *Handler) RecordPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.RecordPayment(c.Request.Context(), id, req.Amount, domain.PaymentMethod(req.PaymentMethod), req.Reference)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment recorded successfully"})
}

func (h *Handler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := orderports.ListOrdersFilter{
		Page:  page,
		Limit: limit,
	}

	if tID := c.Query("table_id"); tID != "" {
		if id, err := strconv.ParseUint(tID, 10, 64); err == nil {
			filter.TableID = &id
		}
	}
	if st := c.Query("status"); st != "" {
		s := domain.Status(st)
		filter.Status = &s
	}

	orders, total, err := h.service.ListOrders(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}

	resp := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, ToOrderResponse(o))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrEmptyOrderItems),
		errors.Is(err, domain.ErrInvalidTableID),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrCannotCompleteUnpaidOrder),
		errors.Is(err, domain.ErrProductNotActive),
		errors.Is(err, domain.ErrActivePriceNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrOptimisticLockConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		// برگرداندن متن واقعی خطا در محیط تست برای خطایابی سریع‌تر
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
