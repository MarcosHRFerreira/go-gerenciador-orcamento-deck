package lossreason

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	lossreasonservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/lossreason"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   lossreasonservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service lossreasonservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/loss-reasons")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)
	protectedRoutes.GET("/:reason_id", h.GetByID)

	adminRoutes := h.router.Group("/loss-reasons")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:reason_id", h.Update)
	adminRoutes.DELETE("/:reason_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateLossReasonRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	id, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	reasonID, ok := parseReasonID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), reasonID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	reasonID, ok := parseReasonID(c)
	if !ok {
		return
	}

	var req dto.UpdateLossReasonRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), reasonID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	reasonID, ok := parseReasonID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), reasonID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseReasonID(c *gin.Context) (int64, bool) {
	rawID := c.Param("reason_id")
	reasonID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || reasonID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "reason_id must be a valid integer")
		return 0, false
	}

	return reasonID, true
}
