package estimator

import (
	"net/http"
	"strconv"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/httpresponse"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	estimatorservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/estimator"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router    gin.IRouter
	validate  *validator.Validate
	service   estimatorservice.Service
	secretKey string
}

func NewHandler(router gin.IRouter, validate *validator.Validate, service estimatorservice.Service, secretKey string) *Handler {
	return &Handler{
		router:    router,
		validate:  validate,
		service:   service,
		secretKey: secretKey,
	}
}

func (h *Handler) RouteList() {
	protectedRoutes := h.router.Group("/estimators")
	protectedRoutes.Use(middleware.Auth(h.secretKey))
	protectedRoutes.GET("", h.List)

	adminRoutes := h.router.Group("/estimators")
	adminRoutes.Use(middleware.Auth(h.secretKey))
	adminRoutes.Use(middleware.RequireRoles(model.RoleAdmin))
	adminRoutes.GET("/next-code", h.GetNextCode)
	adminRoutes.GET("/:estimator_id", h.GetByID)
	adminRoutes.POST("", h.Create)
	adminRoutes.PUT("/:estimator_id", h.Update)
	adminRoutes.DELETE("/:estimator_id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateEstimatorRequest
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

func (h *Handler) GetNextCode(c *gin.Context) {
	code, err := h.service.GetNextCode(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": code})
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
	estimatorID, ok := parseEstimatorID(c)
	if !ok {
		return
	}

	item, err := h.service.GetByID(c.Request.Context(), estimatorID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	estimatorID, ok := parseEstimatorID(c)
	if !ok {
		return
	}

	var req dto.UpdateEstimatorRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	if err := h.service.Update(c.Request.Context(), estimatorID, &req); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	estimatorID, ok := parseEstimatorID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), estimatorID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseEstimatorID(c *gin.Context) (int64, bool) {
	rawID := c.Param("estimator_id")
	estimatorID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || estimatorID <= 0 {
		httpresponse.JSONError(c, http.StatusBadRequest, "estimator_id deve ser um inteiro valido")
		return 0, false
	}

	return estimatorID, true
}
