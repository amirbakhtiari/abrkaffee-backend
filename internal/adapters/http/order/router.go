package order

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("", h.ListOrders)
		orders.GET("/:id", h.GetOrderByID)
		orders.PATCH("/:id/status", h.UpdateStatus)
		orders.POST("/:id/payments", h.RecordPayment)
	}
}
