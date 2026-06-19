package contact

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	contactservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/contact"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   contactservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service contactservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/contacts")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)

	adminRoutes := h.router.Group("/contacts")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.GET("/:contact_id", h.GetByID)
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:contact_id", h.Update)
	adminRoutes.DELETE("/:contact_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateContactRequest
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
	items, err := h.service.List(c.Request.Context(), c.Query("installer_id"))
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetByID(c *gin.Context) {
	contactID, ok := parseContactID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), contactID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	contactID, ok := parseContactID(c)
	if !ok {
		return
	}

	var req dto.UpdateContactRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), contactID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	contactID, ok := parseContactID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), contactID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseContactID(c *gin.Context) (int64, bool) {
	rawID := c.Param("contact_id")
	contactID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || contactID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "contact_id deve ser um inteiro valido")
		return 0, false
	}

	return contactID, true
}
