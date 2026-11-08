package pricing

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domain "github.com/amirbakhtiari/abrcoffee/internal/core/domain/pricing"
	pricingport "github.com/amirbakhtiari/abrcoffee/internal/core/ports/pricing"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service pricingport.PricingService
}

func NewHandler(service pricingport.PricingService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateAnnouncement(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registeredAt := time.Now()
	if req.RegisteredAt != nil {
		registeredAt = *req.RegisteredAt
	}

	items := make([]pricingport.CreatePricingItemDTO, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, pricingport.CreatePricingItemDTO{
			ProductID:          it.ProductID,
			Price:              it.Price,
			AllowPriceChange:   it.AllowPriceChange,
			MaxDiscountPercent: it.MaxDiscountPercent,
		})
	}

	result, err := h.service.CreateAnnouncement(c.Request.Context(), pricingport.CreateAnnouncementDTO{
		Title:        req.Title,
		Code:         req.Code,
		RegisteredAt: registeredAt,
		Items:        items,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"announcement": ToResponse(result)})
}

func (h *Handler) GetAnnouncementByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	result, err := h.service.GetAnnouncementByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"announcement": ToResponse(result)})
}

func (h *Handler) ApproveAnnouncement(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.ApproveAnnouncement(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "price announcement approved successfully"})
}

func (h *Handler) RejectAnnouncement(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.RejectAnnouncement(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "price announcement rejected successfully"})
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid announcement id"})
		return 0, false
	}
	return id, true
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAnnouncementNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "اعلامیه قیمت مورد نظر یافت نشد"})

	case errors.Is(err, domain.ErrProductNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "یک یا چند محصول انتخاب شده در سیستم وجود ندارند"})

	case errors.Is(err, domain.ErrAnnouncementCodeDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "کد اعلامیه قیمت تکراری است"})

	case errors.Is(err, domain.ErrAnnouncementAlreadyHandled):
		c.JSON(http.StatusBadRequest, gin.H{"error": "این اعلامیه قبلاً تعیین وضعیت (تایید یا رد) شده است"})

	case errors.Is(err, domain.ErrDuplicateProductInItems):
		c.JSON(http.StatusBadRequest, gin.H{"error": "یک محصول نمی‌تواند بیش از یک‌بار در اقلام یک اعلامیه تکرار شود"})

	case errors.Is(err, domain.ErrInvalidItemPrice):
		c.JSON(http.StatusBadRequest, gin.H{"error": "قیمت محصول باید بزرگتر از صفر باشد"})

	case errors.Is(err, domain.ErrInvalidDiscountPercent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "درصد تخفیف مجاز باید بین ۰ تا ۱۰۰ باشد"})

	case errors.Is(err, domain.ErrNoItemsProvided):
		c.JSON(http.StatusBadRequest, gin.H{"error": "اعلامیه قیمت باید حداقل شامل یک ردیف محصول باشد"})

	case errors.Is(err, domain.ErrAnnouncementTitleRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "عنوان اعلامیه الزامی است"})

	case errors.Is(err, domain.ErrAnnouncementCodeRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "کد اعلامیه الزامی است"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "خطای داخلی سرور رخ داده است"})
	}
}

func (h *Handler) ListAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	filter := pricingport.ListAnnouncementsFilter{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	items, total, err := h.service.ListAnnouncements(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responseItems := make([]AnnouncementSummaryResponse, 0, len(items))
	for _, it := range items {
		responseItems = append(responseItems, ToSummaryResponse(it))
	}

	c.JSON(http.StatusOK, ListAnnouncementsResponse{
		Items: responseItems,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
