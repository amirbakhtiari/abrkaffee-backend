package pricing

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(admin *gin.RouterGroup, h *Handler) {
	group := admin.Group("/admin/price-announcements")
	{
		group.GET("", h.ListAnnouncements)
		group.POST("", h.CreateAnnouncement)
		group.GET("/:id", h.GetAnnouncementByID)
		group.PATCH("/:id/approve", h.ApproveAnnouncement)
		group.PATCH("/:id/reject", h.RejectAnnouncement)
	}
}
