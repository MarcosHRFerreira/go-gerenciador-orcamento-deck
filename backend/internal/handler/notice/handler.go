package notice

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	noticeservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/notice"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router        gin.IRouter
	validate      *validator.Validate
	noticeService noticeservice.Service
	secretKey     string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, noticeService noticeservice.Service, secretKey string) *Handler {
	return &Handler{
		router:        router,
		validate:      validate,
		noticeService: noticeService,
		secretKey:     secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/notices")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.ListNotices)
	protectedRoutes.GET("/unread-count", h.GetUnreadCount)
	protectedRoutes.GET("/:notice_id", h.GetNotice)
	protectedRoutes.PATCH("/:notice_id/read", h.MarkAsRead)

	adminRoutes := h.router.Group("/notices")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("", h.CreateNotice)
}

func (h *Handler) CreateNotice(c *gin.Context) {
	var req dto.CreateNoticeRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	noticeID, err := h.noticeService.Create(c.Request.Context(), middleware.UserID(c), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateNoticeResponse{ID: noticeID})
}

func (h *Handler) ListNotices(c *gin.Context) {
	statusFilter := strings.TrimSpace(c.Query("status"))
	items, err := h.noticeService.ListByUser(c.Request.Context(), middleware.UserID(c), statusFilter)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	count, err := h.noticeService.CountUnreadByUser(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NoticeUnreadCountResponse{Count: count})
}

func (h *Handler) GetNotice(c *gin.Context) {
	noticeID, ok := parseNoticeID(c)
	if !ok {
		return
	}

	item, err := h.noticeService.GetByIDForUser(c.Request.Context(), middleware.UserID(c), noticeID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	noticeID, ok := parseNoticeID(c)
	if !ok {
		return
	}

	if err := h.noticeService.MarkAsRead(c.Request.Context(), middleware.UserID(c), noticeID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseNoticeID(c *gin.Context) (int64, bool) {
	rawID := strings.TrimSpace(c.Param("notice_id"))
	noticeID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || noticeID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "notice_id deve ser um inteiro valido")
		return 0, false
	}

	return noticeID, true
}
