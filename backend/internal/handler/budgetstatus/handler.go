package budgetstatus

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetstatusservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatus"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   budgetstatusservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service budgetstatusservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/budget-statuses")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)
	protectedRoutes.GET("/:status_id", h.GetByID)

	adminRoutes := h.router.Group("/budget-statuses")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:status_id", h.Update)
	adminRoutes.DELETE("/:status_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateBudgetStatusRequest
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
	statusID, ok := parseStatusID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), statusID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	statusID, ok := parseStatusID(c)
	if !ok {
		return
	}

	var req dto.UpdateBudgetStatusRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), statusID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	statusID, ok := parseStatusID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), statusID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseStatusID(c *gin.Context) (int64, bool) {
	rawID := c.Param("status_id")
	statusID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || statusID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "status_id deve ser um inteiro valido")
		return 0, false
	}

	return statusID, true
}
